package token

import (
	"testing"
)

func TestToken_EncodeAndExtract(t *testing.T) {
	email := "testuser@example.com"
	tok, err := GenerateRegistrationToken(email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	extracted, err := ExtractEmailFromToken(tok)
	if err != nil {
		t.Fatalf("failed to extract email: %v", err)
	}

	if extracted != email {
		t.Errorf("expected email %s, got %s", email, extracted)
	}
}

func TestToken_InvalidToken(t *testing.T) {
	_, err := ExtractEmailFromToken("invalid-base64-random-string")
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}
