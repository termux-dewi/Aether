// aether-server is the AetherL3-Tunnel server daemon.
// It manages a TUN device, accepts gRPC tunnel connections,
// and forwards L3 packets between clients and the internet.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/aether-l3/aether/internal/server"
	"github.com/aether-l3/aether/pkg/auth"
	"github.com/aether-l3/aether/pkg/config"
	"github.com/aether-l3/aether/pkg/tun"
	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	listenAddr := flag.String("listen", ":443", "gRPC listen address")
	psk := flag.String("psk", "", "Pre-shared key (hex-encoded, 64 chars)")
	certFile := flag.String("cert", "", "TLS certificate file")
	keyFile := flag.String("key", "", "TLS private key file")
	tunName := flag.String("tun-name", "aether0", "TUN device name")
	tunCIDR := flag.String("tun-cidr", "10.8.0.1/24", "TUN IPv4 CIDR")
	tunMTU := flag.Int("tun-mtu", 1500, "TUN MTU")
	maxClients := flag.Int("max-clients", 64, "Maximum concurrent clients")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	// Setup logger.
	level := slog.LevelInfo
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	if *psk == "" {
		logger.Error("PSK is required. Generate one with: openssl rand -hex 32")
		os.Exit(1)
	}

	// Validate config.
	cfg := &config.ServerConfig{
		ListenAddr: *listenAddr,
		PSK:        *psk,
		MaxClients: *maxClients,
		TUN: config.TUNConfig{
			DeviceName: *tunName,
			IPv4CIDR:   *tunCIDR,
			MTU:        *tunMTU,
		},
		TLS: config.TLSConfig{
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
	}
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid config", "err", err)
		os.Exit(1)
	}

	// Create TUN device.
	logger.Info("creating TUN device", "name", cfg.TUN.DeviceName, "cidr", cfg.TUN.IPv4CIDR)
	dev, err := tun.CreateTUN(tun.Config{
		Name:     cfg.TUN.DeviceName,
		IPv4CIDR: cfg.TUN.IPv4CIDR,
		MTU:      cfg.TUN.MTU,
	})
	if err != nil {
		logger.Error("creating TUN device", "err", err)
		os.Exit(1)
	}
	defer dev.Close()
	logger.Info("TUN device created", "name", dev.Name())

	// Setup authenticator.
	authenticator, err := auth.NewAuthenticator(cfg.PSK)
	if err != nil {
		logger.Error("creating authenticator", "err", err)
		os.Exit(1)
	}

	// Setup session manager.
	sessions := server.NewSessionManager(cfg.MaxClients, 0x0A080000) // 10.8.0.0

	// Setup packet handler.
	packetHandler := server.NewTUNPacketHandler(dev, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start TUN read loop.
	packetHandler.StartReadLoop(ctx)

	// Create gRPC server.
	var grpcOpts []grpc.ServerOption

	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		tlsCfg, err := cfg.TLS.BuildTLSConfig(true)
		if err != nil {
			logger.Error("building TLS config", "err", err)
			os.Exit(1)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(tlsCfg)))
		logger.Info("TLS enabled", "cert", cfg.TLS.CertFile)
	} else {
		logger.Warn("TLS disabled - running in insecure mode (development only)")
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	tunnelSvc := server.NewTunnelService(authenticator, sessions, packetHandler, packetHandler, logger)
	aetherv1.RegisterTunnelServiceServer(grpcServer, tunnelSvc)

	// Start listening.
	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		logger.Error("listening", "addr", cfg.ListenAddr, "err", err)
		os.Exit(1)
	}

	logger.Info("aether-server starting", "addr", cfg.ListenAddr, "tun", dev.Name())

	// Graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("shutting down...")
		cancel()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC serve", "err", err)
		os.Exit(1)
	}

	fmt.Println("aether-server stopped")
}