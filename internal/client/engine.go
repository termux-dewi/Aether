// Package client implements the AetherL3 tunnel client engine.
// This package is designed to be used both standalone and via gomobile
// bindings in the Android AetherBox application.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aether-l3/aether/pkg/auth"
	"github.com/aether-l3/aether/pkg/frame"
	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// EngineConfig holds client engine configuration.
type EngineConfig struct {
	ServerAddr string
	PSK        string
	ServerName string // TLS SNI
	Insecure   bool   // Skip TLS verify (dev only)

	KeepAliveInterval time.Duration
	ReconnectEnabled  bool
	MaxReconnect      int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
}

// Engine is the tunnel client engine.
type Engine struct {
	cfg    EngineConfig
	logger *slog.Logger

	authenticator *auth.Authenticator
	conn          *grpc.ClientConn
	tunnelClient  aetherv1.TunnelServiceClient
	tunnelStream  aetherv1.TunnelService_TunnelClient

	sessionToken []byte
	tunnelIP     string
	mtu          uint32

	encoder *frame.Encoder
	decoder *frame.Decoder

	outbound chan []byte
	inbound  chan []byte

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	connected bool
}

// NewEngine creates a new tunnel client engine.
func NewEngine(cfg EngineConfig, logger *slog.Logger) (*Engine, error) {
	authenticator, err := auth.NewAuthenticator(cfg.PSK)
	if err != nil {
		return nil, fmt.Errorf("creating authenticator: %w", err)
	}

	if cfg.KeepAliveInterval <= 0 {
		cfg.KeepAliveInterval = 15 * time.Second
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = 1 * time.Second
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		cfg:           cfg,
		logger:        logger,
		authenticator: authenticator,
		outbound:      make(chan []byte, 512),
		inbound:       make(chan []byte, 512),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Connect establishes the gRPC connection and performs handshake.
func (e *Engine) Connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.connected {
		return fmt.Errorf("already connected")
	}

	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: e.cfg.Insecure,
	}
	if e.cfg.ServerName != "" {
		tlsCfg.ServerName = e.cfg.ServerName
	}

	var opts []grpc.DialOption
	if !e.cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.NewClient(e.cfg.ServerAddr, opts...)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", e.cfg.ServerAddr, err)
	}

	e.conn = conn
	e.tunnelClient = aetherv1.NewTunnelServiceClient(conn)

	token := e.authenticator.GenerateSimpleToken()
	resp, err := e.tunnelClient.Handshake(e.ctx, &aetherv1.HandshakeRequest{
		ProtocolVersion: frame.ProtocolVersion,
		AuthToken:       token,
		ClientId:        "AetherBox/1.0",
		RequestedMtu:    1500,
	})
	if err != nil {
		conn.Close()
		return fmt.Errorf("handshake: %w", err)
	}

	if !resp.Accepted {
		conn.Close()
		return fmt.Errorf("handshake rejected: %s", resp.Reason)
	}

	e.sessionToken = resp.SessionToken
	e.tunnelIP = resp.TunnelIp
	e.mtu = resp.NegotiatedMtu
	e.encoder = frame.NewEncoder(e.sessionToken)
	e.decoder = frame.NewDecoder(e.sessionToken)

	stream, err := e.tunnelClient.Tunnel(e.ctx)
	if err != nil {
		conn.Close()
		return fmt.Errorf("opening tunnel stream: %w", err)
	}

	initFrame := e.encoder.EncodeKeepAlive()
	if err := stream.Send(initFrame); err != nil {
		conn.Close()
		return fmt.Errorf("sending init frame: %w", err)
	}

	e.tunnelStream = stream
	e.connected = true

	e.logger.Info("connected", "server", e.cfg.ServerAddr, "tunnel_ip", e.tunnelIP, "mtu", e.mtu)

	go e.pumpOutbound()
	go e.pumpInbound()
	go e.keepAliveLoop()

	return nil
}