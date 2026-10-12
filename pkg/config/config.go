// Package config defines the configuration structures for AetherL3-Tunnel.
package config

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// TransportMode selects the underlying transport protocol.
type TransportMode string

const (
	TransportGRPC  TransportMode = "grpc"  // gRPC over HTTP/2
	TransportHTTP3 TransportMode = "http3" // HTTP/3 over QUIC
)

// ServerConfig holds all server-side configuration.
type ServerConfig struct {
	// Listen address for the gRPC/HTTP3 server.
	ListenAddr string `json:"listen_addr" yaml:"listen_addr"`

	// Transport mode.
	Transport TransportMode `json:"transport" yaml:"transport"`

	// TLS configuration.
	TLS TLSConfig `json:"tls" yaml:"tls"`

	// Pre-Shared Key for authentication (hex-encoded).
	PSK string `json:"psk" yaml:"psk"`

	// TUN device configuration.
	TUN TUNConfig `json:"tun" yaml:"tun"`

	// Maximum number of concurrent client sessions.
	MaxClients int `json:"max_clients" yaml:"max_clients"`

	// Keep-alive interval for tunnel health checks.
	KeepAliveInterval time.Duration `json:"keepalive_interval" yaml:"keepalive_interval"`

	// Keep-alive timeout before declaring session dead.
	KeepAliveTimeout time.Duration `json:"keepalive_timeout" yaml:"keepalive_timeout"`

	// Enable IPv6 support.
	EnableIPv6 bool `json:"enable_ipv6" yaml:"enable_ipv6"`

	// Log level: debug, info, warn, error.
	LogLevel string `json:"log_level" yaml:"log_level"`
}

// TUNConfig holds TUN interface configuration.
type TUNConfig struct {
	// Interface name (e.g., "aether0").
	DeviceName string `json:"device_name" yaml:"device_name"`

	// IPv4 address and CIDR for the TUN interface (e.g., "10.8.0.1/24").
	IPv4CIDR string `json:"ipv4_cidr" yaml:"ipv4_cidr"`

	// IPv6 address and CIDR for the TUN interface (e.g., "fd00::1/64").
	IPv6CIDR string `json:"ipv6_cidr" yaml:"ipv6_cidr"`

	// MTU for the TUN interface.
	MTU int `json:"mtu" yaml:"mtu"`
}

// TLSConfig holds TLS certificate configuration.
type TLSConfig struct {
	// Path to server certificate (PEM).
	CertFile string `json:"cert_file" yaml:"cert_file"`

	// Path to server private key (PEM).
	KeyFile string `json:"key_file" yaml:"key_file"`

	// Path to CA certificate for client verification (optional).
	CAFile string `json:"ca_file" yaml:"ca_file"`

	// Server name for TLS SNI.
	ServerName string `json:"server_name" yaml:"server_name"`

	// Skip certificate verification (insecure, for testing only).
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// ClientConfig holds all client-side configuration.
type ClientConfig struct {
	// Server address (host:port).
	ServerAddr string `json:"server_addr" yaml:"server_addr"`

	// Transport mode.
	Transport TransportMode `json:"transport" yaml:"transport"`

	// TLS configuration.
	TLS TLSConfig `json:"tls" yaml:"tls"`

	// Pre-Shared Key for authentication (hex-encoded).
	PSK string `json:"psk" yaml:"psk"`

	// TUN device configuration (client side, for Android VpnService or desktop).
	TUN TUNConfig `json:"tun" yaml:"tun"`

	// Keep-alive interval.
	KeepAliveInterval time.Duration `json:"keepalive_interval" yaml:"keepalive_interval"`

	// Reconnect settings.
	Reconnect ReconnectConfig `json:"reconnect" yaml:"reconnect"`

	// Split tunneling: list of allowed package names (Android only).
	AllowedApps []string `json:"allowed_apps" yaml:"allowed_apps"`

	// Split tunneling: list of disallowed package names (Android only).
	DisallowedApps []string `json:"disallowed_apps" yaml:"disallowed_apps"`

	// Log level.
	LogLevel string `json:"log_level" yaml:"log_level"`
}

// ReconnectConfig holds auto-reconnect settings.
type ReconnectConfig struct {
	// Enable automatic reconnection.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Maximum number of reconnect attempts (0 = unlimited).
	MaxAttempts int `json:"max_attempts" yaml:"max_attempts"`

	// Initial backoff duration.
	InitialBackoff time.Duration `json:"initial_backoff" yaml:"initial_backoff"`

	// Maximum backoff duration.
	MaxBackoff time.Duration `json:"max_backoff" yaml:"max_backoff"`

	// Backoff multiplier.
	BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`
}

// Validate checks the server configuration for errors.
func (c *ServerConfig) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen_addr is required")
	}
	if c.PSK == "" {
		return fmt.Errorf("psk is required")
	}
	if c.TUN.DeviceName == "" {
		return fmt.Errorf("tun.device_name is required")
	}
	if c.TUN.IPv4CIDR == "" {
		return fmt.Errorf("tun.ipv4_cidr is required")
	}
	if _, _, err := net.ParseCIDR(c.TUN.IPv4CIDR); err != nil {
		return fmt.Errorf("invalid tun.ipv4_cidr: %w", err)
	}
	if c.TUN.MTU <= 0 {
		c.TUN.MTU = 1500
	}
	if c.MaxClients <= 0 {
		c.MaxClients = 64
	}
	if c.KeepAliveInterval <= 0 {
		c.KeepAliveInterval = 15 * time.Second
	}
	if c.KeepAliveTimeout <= 0 {
		c.KeepAliveTimeout = 45 * time.Second
	}
	return nil
}

// Validate checks the client configuration for errors.
func (c *ClientConfig) Validate() error {
	if c.ServerAddr == "" {
		return fmt.Errorf("server_addr is required")
	}
	if c.PSK == "" {
		return fmt.Errorf("psk is required")
	}
	if c.KeepAliveInterval <= 0 {
		c.KeepAliveInterval = 15 * time.Second
	}
	return nil
}

// BuildTLSConfig constructs a *tls.Config from TLSConfig.
func (tc *TLSConfig) BuildTLSConfig(isServer bool) (*tls.Config, error) {
	config := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: tc.InsecureSkipVerify,
	}

	if tc.ServerName != "" {
		config.ServerName = tc.ServerName
	}

	if isServer {
		if tc.CertFile == "" || tc.KeyFile == "" {
			return nil, fmt.Errorf("server requires cert_file and key_file")
		}
		cert, err := tls.LoadX509KeyPair(tc.CertFile, tc.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading server certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	return config, nil
}