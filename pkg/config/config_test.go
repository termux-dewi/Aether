package config

import (
	"testing"
)

func TestServerConfigValidate(t *testing.T) {
	cfg := &ServerConfig{
		ListenAddr: ":443",
		PSK:        "abcd1234abcd1234abcd1234abcd1234",
		TUN: TUNConfig{
			DeviceName: "aether0",
			IPv4CIDR:   "10.8.0.1/24",
			MTU:        1400,
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if cfg.MaxClients != 64 {
		t.Fatalf("expected default MaxClients 64, got %d", cfg.MaxClients)
	}
}

func TestServerConfigValidateMissingPSK(t *testing.T) {
	cfg := &ServerConfig{
		ListenAddr: ":443",
		TUN:        TUNConfig{DeviceName: "tun0", IPv4CIDR: "10.0.0.1/24"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing PSK")
	}
}

func TestServerConfigValidateInvalidCIDR(t *testing.T) {
	cfg := &ServerConfig{
		ListenAddr: ":443",
		PSK:        "abcd1234abcd1234abcd1234abcd1234",
		TUN:        TUNConfig{DeviceName: "tun0", IPv4CIDR: "invalid"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}

func TestClientConfigValidate(t *testing.T) {
	cfg := &ClientConfig{
		ServerAddr: "example.com:443",
		PSK:        "abcd1234abcd1234abcd1234abcd1234",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
}

func TestClientConfigValidateMissing(t *testing.T) {
	cfg := &ClientConfig{PSK: "abcd1234abcd1234abcd1234abcd1234"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing server_addr")
	}
}