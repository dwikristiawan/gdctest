package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid registration token format")
)

// GenerateRegistrationToken encodes email, timestamp, and random bytes into a URL-safe string
func GenerateRegistrationToken(email string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	randomPart := base64.RawURLEncoding.EncodeToString(b)
	ts := time.Now().Unix()

	raw := fmt.Sprintf("%s|%d|%s", email, ts, randomPart)
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

// ExtractEmailFromToken decodes email from the token payload
func ExtractEmailFromToken(tok string) (string, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(tok)
	if err != nil {
		return "", ErrInvalidToken
	}

	parts := strings.Split(string(decodedBytes), "|")
	if len(parts) < 3 {
		return "", ErrInvalidToken
	}

	email := strings.TrimSpace(parts[0])
	if email == "" {
		return "", ErrInvalidToken
	}

	return email, nil
}
