// Package mobile provides gomobile-compatible bindings for the AetherL3
// tunnel client engine, designed to be consumed from Android (Kotlin/Java)
// via gomobile bind.
package mobile

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aether-l3/aether/internal/client"
)

// TunnelEngine is the gomobile-exposed tunnel engine.
type TunnelEngine struct {
	engine *client.Engine
	logger *slog.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// TunnelConfig holds configuration passed from the Android app.
type TunnelConfig struct {
	ServerAddr string
	PSK        string
	ServerName string
	Insecure   bool
}

// NewTunnelEngine creates a new tunnel engine for Android.
func NewTunnelEngine(cfg *TunnelConfig) (*TunnelEngine, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	engineCfg := client.EngineConfig{
		ServerAddr:        cfg.ServerAddr,
		PSK:               cfg.PSK,
		ServerName:        cfg.ServerName,
		Insecure:          cfg.Insecure,
		KeepAliveInterval: 15 * time.Second,
		ReconnectEnabled:  true,
		MaxReconnect:      0,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
	}

	engine, err := client.NewEngine(engineCfg, logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TunnelEngine{
		engine: engine,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Connect establishes the tunnel connection.
func (t *TunnelEngine) Connect() error {
	return t.engine.Connect()
}

// Disconnect closes the tunnel.
func (t *TunnelEngine) Disconnect() {
	t.engine.Disconnect()
	t.cancel()
}

// IsConnected returns whether the tunnel is active.
func (t *TunnelEngine) IsConnected() bool {
	return t.engine.IsConnected()
}

// GetTunnelIP returns the assigned tunnel IP address.
func (t *TunnelEngine) GetTunnelIP() string {
	return t.engine.TunnelIP()
}

// GetMTU returns the negotiated MTU.
func (t *TunnelEngine) GetMTU() int32 {
	return int32(t.engine.MTU())
}

// WritePacket sends a raw IP packet into the tunnel.
// Called from Android VpnService when packets are read from the TUN fd.
func (t *TunnelEngine) WritePacket(packet []byte) {
	t.engine.WritePacket(packet)
}

// ReadPacket blocks until a packet is available from the tunnel.
// Returns nil if context is cancelled.
func (t *TunnelEngine) ReadPacket() []byte {
	pkt, err := t.engine.ReadPacket(t.ctx)
	if err != nil {
		return nil
	}
	return pkt
}