package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MercadoPago only sends credentials to the provider's fixed API origin.
type MercadoPago struct {
	Token  string
	Client *http.Client
}
type MPPreference struct {
	ID               string `json:"id"`
	InitPoint        string `json:"init_point"`
	SandboxInitPoint string `json:"sandbox_init_point"`
}
type MPPayment struct {
	ID                int64       `json:"id"`
	Status            string      `json:"status"`
	ExternalReference string      `json:"external_reference"`
	CurrencyID        string      `json:"currency_id"`
	TransactionAmount json.Number `json:"transaction_amount"`
	AmountRefunded    json.Number `json:"transaction_amount_refunded"`
	CollectorID       int64       `json:"collector_id"`
	LiveMode          bool        `json:"live_mode"`
}

type MPUser struct {
	ID   int64    `json:"id"`
	Tags []string `json:"tags"`
}

// Test sellers can return live_mode=true. Confirm the token owner is an
// actual provider test user before accepting that combination.
func ValidMPEnvironment(testMode, liveMode bool, collectorID int64, seller MPUser) bool {
	if !testMode {
		return liveMode
	}
	if !liveMode {
		return true
	}
	if seller.ID != collectorID {
		return false
	}
	for _, tag := range seller.Tags {
		if tag == "test_user" {
			return true
		}
	}
	return false
}

func (m MercadoPago) Request(ctx context.Context, method, path string, body any, output any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.mercadopago.com"+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.Token)
	req.Header.Set("Content-Type", "application/json")
	client := m.Client
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("mercado pago unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mercado pago HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(output)
}

// Parse decimal pesos exactly; never round an untrusted provider amount.
func AmountCents(value json.Number) (int64, error) {
	parts := strings.Split(string(value), ".")
	if len(parts) > 2 || len(parts) == 0 || parts[0] == "" || strings.HasPrefix(parts[0], "-") {
		return 0, errors.New("invalid amount")
	}
	fraction := "00"
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) > 2 {
			return 0, errors.New("invalid precision")
		}
		fraction += strings.Repeat("0", 2-len(fraction))
	}
	for _, c := range parts[0] + fraction {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid amount")
		}
	}
	return strconv.ParseInt(parts[0]+fraction, 10, 64)
}

func ValidMPSignature(signature, requestID, dataID, secret string) bool {
	if secret == "" || requestID == "" || dataID == "" {
		return false
	}
	var ts, digest string
	for _, part := range strings.Split(signature, ",") {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			return false
		}
		switch pair[0] {
		case "ts":
			ts = pair[1]
		case "v1":
			digest = pair[1]
		}
	}
	if _, err := strconv.ParseInt(ts, 10, 64); err != nil {
		return false
	}
	expected, err := hex.DecodeString(digest)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "id:%s;request-id:%s;ts:%s;", strings.ToLower(dataID), requestID, ts)
	return hmac.Equal(expected, mac.Sum(nil))
}
