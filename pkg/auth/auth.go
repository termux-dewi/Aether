// Package auth provides PSK-based authentication for AetherL3-Tunnel.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// TokenLength is the length of the generated session token in bytes.
	TokenLength = 32

	// NonceLength is the length of the challenge nonce in bytes.
	NonceLength = 16

	// MaxTokenAge is the maximum age of a valid authentication token.
	MaxTokenAge = 5 * time.Minute
)

// Authenticator handles PSK-based authentication.
type Authenticator struct {
	psk []byte
}

// NewAuthenticator creates a new Authenticator from a hex-encoded PSK string.
func NewAuthenticator(hexPSK string) (*Authenticator, error) {
	psk, err := hex.DecodeString(hexPSK)
	if err != nil {
		return nil, fmt.Errorf("invalid PSK hex encoding: %w", err)
	}
	if len(psk) < 16 {
		return nil, fmt.Errorf("PSK must be at least 16 bytes (32 hex chars)")
	}
	return &Authenticator{psk: psk}, nil
}

// GenerateChallenge creates a random nonce for challenge-response auth.
func (a *Authenticator) GenerateChallenge() ([]byte, error) {
	nonce := make([]byte, NonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating challenge nonce: %w", err)
	}
	return nonce, nil
}

// ComputeToken computes the HMAC-SHA256 authentication token.
// The token is computed over: nonce || timestamp_bytes.
func (a *Authenticator) ComputeToken(nonce []byte, timestamp time.Time) []byte {
	mac := hmac.New(sha256.New, a.psk)
	mac.Write(nonce)
	ts := timestamp.UnixNano()
	tsBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		tsBytes[i] = byte(ts)
		ts >>= 8
	}
	mac.Write(tsBytes)
	return mac.Sum(nil)
}

// VerifyToken verifies a client-provided authentication token.
// Returns nil if valid, error otherwise.
func (a *Authenticator) VerifyToken(nonce []byte, timestamp time.Time, token []byte) error {
	// Check token age.
	age := time.Since(timestamp)
	if age < 0 {
		age = -age
	}
	if age > MaxTokenAge {
		return fmt.Errorf("token expired (age: %v, max: %v)", age, MaxTokenAge)
	}

	expected := a.ComputeToken(nonce, timestamp)
	if !hmac.Equal(expected, token) {
		return fmt.Errorf("authentication failed: invalid token")
	}
	return nil
}

// VerifySimpleToken verifies a token computed directly from PSK (simplified auth).
// Used for initial connection where client sends HMAC(psk, "aether-auth-v1").
func (a *Authenticator) VerifySimpleToken(token []byte) error {
	mac := hmac.New(sha256.New, a.psk)
	mac.Write([]byte("aether-auth-v1"))
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, token) {
		return fmt.Errorf("authentication failed: invalid simple token")
	}
	return nil
}

// GenerateSimpleToken generates the simplified auth token.
func (a *Authenticator) GenerateSimpleToken() []byte {
	mac := hmac.New(sha256.New, a.psk)
	mac.Write([]byte("aether-auth-v1"))
	return mac.Sum(nil)
}

// GenerateSessionToken creates a random session token for an authenticated session.
func GenerateSessionToken() ([]byte, error) {
	token := make([]byte, TokenLength)
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}
	return token, nil
}

// GeneratePSK generates a new random PSK and returns it hex-encoded.
func GeneratePSK() (string, error) {
	psk := make([]byte, 32)
	if _, err := rand.Read(psk); err != nil {
		return "", fmt.Errorf("generating PSK: %w", err)
	}
	return hex.EncodeToString(psk), nil
}