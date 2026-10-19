package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAmountCents(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int64
		valid bool
	}{
		{"1234.56", 123456, true}, {"1234.5", 123450, true}, {"0", 0, true}, {"10", 1000, true},
		{"-1", 0, false}, {"1.001", 0, false}, {"1e2", 0, false}, {"", 0, false}, {"99999999999999999999", 0, false},
	} {
		got, err := AmountCents(json.Number(tc.value))
		if (err == nil) != tc.valid || (tc.valid && got != tc.want) {
			t.Errorf("%s: got %d, %v", tc.value, got, err)
		}
	}
}

func TestMPEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name               string
		testMode, liveMode bool
		seller             MPUser
		want               bool
	}{
		{"sandbox", true, false, MPUser{}, true},
		{"test seller live flag", true, true, MPUser{ID: 42, Tags: []string{"test_user"}}, true},
		{"real seller blocked in test", true, true, MPUser{ID: 42, Tags: []string{"normal"}}, false},
		{"wrong test seller", true, true, MPUser{ID: 99, Tags: []string{"test_user"}}, false},
		{"production", false, true, MPUser{}, true},
		{"sandbox blocked in production", false, false, MPUser{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidMPEnvironment(tc.testMode, tc.liveMode, 42, tc.seller); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestMPSignature(t *testing.T) {
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte("id:12345;request-id:request-1;ts:1704908010;"))
	signature := "ts=1704908010,v1=" + hex.EncodeToString(mac.Sum(nil))
	if !ValidMPSignature(signature, "request-1", "12345", "test-secret") {
		t.Fatal("valid signature rejected")
	}
	for _, tc := range []struct{ sig, request, id, secret string }{
		{signature, "request-1", "999", "test-secret"}, {signature, "other", "12345", "test-secret"},
		{signature, "request-1", "12345", "wrong"}, {signature, "request-1", "12345", ""},
		{"ts=no,v1=00", "request-1", "12345", "test-secret"}, {"", "request-1", "12345", "test-secret"},
	} {
		if ValidMPSignature(tc.sig, tc.request, tc.id, tc.secret) {
			t.Fatal("invalid signature accepted")
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestMPRequest(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.mercadopago.com/v1/payments/123" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("incorrect API request")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":123,"transaction_amount":1234.56}`)), Header: make(http.Header)}, nil
	})}
	var payment MPPayment
	err := (MercadoPago{Token: "secret", Client: client}).Request(context.Background(), http.MethodGet, "/v1/payments/123", nil, &payment)
	if err != nil || payment.ID != 123 || payment.TransactionAmount != "1234.56" {
		t.Fatalf("unexpected payment: %+v %v", payment, err)
	}
}
