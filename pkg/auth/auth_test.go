package auth

import (
	"testing"
	"time"
)

func TestGeneratePSK(t *testing.T) {
	psk, err := GeneratePSK()
	if err != nil {
		t.Fatalf("GeneratePSK() error: %v", err)
	}
	if len(psk) != 64 { // 32 bytes = 64 hex chars
		t.Fatalf("expected 64 hex chars, got %d", len(psk))
	}

	// Generate again should be different
	psk2, _ := GeneratePSK()
	if psk == psk2 {
		t.Fatal("two generated PSKs should differ")
	}
}

func TestNewAuthenticator(t *testing.T) {
	psk, _ := GeneratePSK()
	a, err := NewAuthenticator(psk)
	if err != nil {
		t.Fatalf("NewAuthenticator() error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil authenticator")
	}

	// Invalid hex
	_, err = NewAuthenticator("not-hex")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}

	// Too short PSK
	_, err = NewAuthenticator("abcd")
	if err == nil {
		t.Fatal("expected error for short PSK")
	}
}

func TestSimpleTokenAuth(t *testing.T) {
	psk, _ := GeneratePSK()
	a, _ := NewAuthenticator(psk)

	token := a.GenerateSimpleToken()
	if err := a.VerifySimpleToken(token); err != nil {
		t.Fatalf("VerifySimpleToken() error: %v", err)
	}

	// Wrong token should fail
	wrong := make([]byte, len(token))
	copy(wrong, token)
	wrong[0] ^= 0xFF
	if err := a.VerifySimpleToken(wrong); err == nil {
		t.Fatal("expected error for wrong token")
	}
}

func TestChallengeResponseAuth(t *testing.T) {
	psk, _ := GeneratePSK()
	a, _ := NewAuthenticator(psk)

	nonce, err := a.GenerateChallenge()
	if err != nil {
		t.Fatalf("GenerateChallenge() error: %v", err)
	}

	ts := time.Now()
	token := a.ComputeToken(nonce, ts)

	if err := a.VerifyToken(nonce, ts, token); err != nil {
		t.Fatalf("VerifyToken() error: %v", err)
	}

	// Expired token
	oldTs := time.Now().Add(-10 * time.Minute)
	oldToken := a.ComputeToken(nonce, oldTs)
	if err := a.VerifyToken(nonce, oldTs, oldToken); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken() error: %v", err)
	}
	if len(token) != TokenLength {
		t.Fatalf("expected %d bytes, got %d", TokenLength, len(token))
	}
}