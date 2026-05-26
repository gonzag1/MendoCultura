package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID uint   `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

func GenerateJWT(claims Claims, secret string) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := sign(unsigned, secret)
	return unsigned + "." + signature, nil
}

func ParseJWT(token string, secret string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("invalid token format")
	}

	unsigned := parts[0] + "." + parts[1]
	expected := sign(unsigned, secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return Claims{}, errors.New("invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, err
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, err
	}
	if claims.Exp < time.Now().Unix() {
		return Claims{}, errors.New("expired token")
	}

	return claims, nil
}

func NewTicketCode(ticketID uint, secret string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	raw := fmt.Sprintf("%d.%s", ticketID, hex.EncodeToString(randomBytes))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(raw))
	signature := hex.EncodeToString(mac.Sum(nil))[:16]
	return "MC-" + base64.RawURLEncoding.EncodeToString([]byte(raw+"."+signature)), nil
}

func ValidateTicketCode(code string, secret string) (uint, error) {
	if !strings.HasPrefix(code, "MC-") {
		return 0, errors.New("invalid code prefix")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(code, "MC-"))
	if err != nil {
		return 0, err
	}

	parts := strings.Split(string(decoded), ".")
	if len(parts) != 3 {
		return 0, errors.New("invalid code payload")
	}

	raw := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(raw))
	expected := hex.EncodeToString(mac.Sum(nil))[:16]
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return 0, errors.New("invalid code signature")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func sign(unsigned string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
