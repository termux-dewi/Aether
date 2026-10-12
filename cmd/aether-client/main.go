// aether-client is a standalone desktop tunnel client for testing.
// On Android, the same internal/client engine is used via gomobile bindings.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aether-l3/aether/internal/client"
	"github.com/aether-l3/aether/pkg/tun"
)

func main() {
	serverAddr := flag.String("server", "127.0.0.1:443", "Server address (host:port)")
	psk := flag.String("psk", "", "Pre-shared key (hex-encoded)")
	sni := flag.String("sni", "", "TLS SNI server name")
	insecure := flag.Bool("insecure", false, "Skip TLS verification (dev only)")
	tunName := flag.String("tun-name", "aether-tun0", "Client TUN device name")
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

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
		logger.Error("PSK is required (--psk flag)")
		os.Exit(1)
	}

	cfg := client.EngineConfig{
		ServerAddr:        *serverAddr,
		PSK:               *psk,
		ServerName:        *sni,
		Insecure:          *insecure,
		KeepAliveInterval: 15 * time.Second,
		ReconnectEnabled:  true,
		MaxReconnect:      0, // unlimited
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
	}

	engine, err := client.NewEngine(cfg, logger)
	if err != nil {
		logger.Error("creating engine", "err", err)
		os.Exit(1)
	}

	if err := engine.Connect(); err != nil {
		logger.Error("connecting", "err", err)
		os.Exit(1)
	}

	defer engine.Disconnect()

	// Create client TUN device.
	dev, err := tun.CreateTUN(tun.Config{
		Name: *tunName,
		MTU:  int(engine.MTU()),
	})
	if err != nil {
		logger.Error("creating TUN device", "err", err)
		os.Exit(1)
	}
	defer dev.Close()

	// Configure TUN with assigned tunnel IP.
	tunnelIP := engine.TunnelIP()
	logger.Info("configuring TUN", "name", dev.Name(), "ip", tunnelIP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TUN -> tunnel
	go func() {
		buf := make([]byte, 65535)
		for {
			n, err := dev.Read(buf)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Error("reading TUN", "err", err)
				continue
			}
			engine.WritePacket(buf[:n])
		}
	}()

	// tunnel -> TUN
	go func() {
		for {
			pkt, err := engine.ReadPacket(ctx)
			if err != nil {
				return
			}
			if _, err := dev.Write(pkt); err != nil {
				logger.Error("writing TUN", "err", err)
			}
		}
	}()

	logger.Info("aether-client running", "tun", dev.Name(), "tunnel_ip", tunnelIP)

	// Wait for signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("shutting down")
}