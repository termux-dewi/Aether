# AetherL3-Tunnel

Layer 3 tunneling system over gRPC (HTTP/2) with TLS 1.3, supporting raw ICMP, TCP, and UDP IP packet forwarding through a single encrypted connection.

## Architecture

```
[Android Client / Desktop Client]
       │ VpnService TUN / tun device
       ▼
[Client Engine (Go)]  ──── gRPC bidi stream ────►  [Server Daemon]
       │ encapsulated AetherFrame                      │
       ▼                                               ▼
   IP packets                                    TUN device → NAT → Internet
```

## Components

| Component | Path | Description |
|-----------|------|-------------|
| Server daemon | `cmd/aether-server` | Runs on VPS, manages TUN + NAT |
| Desktop client | `cmd/aether-client` | Standalone client for testing |
| Android client | `android/` | AetherBox (Kotlin + Jetpack Compose) |
| Proto definition | `proto/aether/v1/tunnel.proto` | gRPC service + message schema |
| Core packages | `pkg/` | auth, config, frame, tun |
| Server internals | `internal/server/` | session mgmt, gRPC service, TUN bridge |
| Client internals | `internal/client/` | tunnel engine (gomobile-compatible) |

## Quick Start

### Build

```bash
make build        # builds both server and client binaries
make test         # runs all unit tests
make proto        # regenerates protobuf code
```

### Server Setup (Linux VPS)

```bash
# 1. Generate a PSK
openssl rand -hex 32

# 2. Run the server (requires root for TUN)
sudo ./bin/aether-server \
  -listen :443 \
  -psk <your-psk> \
  -cert /path/to/cert.pem \
  -key /path/to/key.pem \
  -tun-name aether0 \
  -tun-cidr 10.8.0.1/24

# 3. Enable NAT forwarding
sudo ./cmd/aether-server/setup_nat.sh aether0 eth0
```

### Client (Desktop / Testing)

```bash
sudo ./bin/aether-client \
  -server your-vps:443 \
  -psk <your-psk> \
  -sni your-domain.com

# Then test:
ping -I aether-tun0 8.8.8.8
```

### Android (AetherBox)

See `android/README.md` for build instructions with Android Studio + gomobile bindings.

## Protocol

- Transport: gRPC over TLS 1.3 (HTTP/2 multiplexed streams)
- Authentication: PSK-derived HMAC-SHA256 token
- Framing: `AetherFrame` protobuf messages carrying raw IPv4/IPv6 packets
- Multiplexing: independent bidirectional gRPC stream per session
- Keep-alive: periodic `KEEPALIVE`/`KEEPALIVE_ACK` frames with expiry-based session cleanup

## Security Notes

- PSK must be ≥ 32 bytes (64 hex characters)
- TLS 1.3 minimum enforced; TLS 1.2 and below rejected
- Server requires valid certificate in production (insecure mode is dev-only)

## License

MIT