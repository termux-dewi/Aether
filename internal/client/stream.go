package client

import (
	"context"
	"io"
	"time"

	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
)

// pumpOutbound reads packets from the outbound channel and sends to server.
func (e *Engine) pumpOutbound() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case pkt := <-e.outbound:
			f, err := e.encoder.EncodeData(pkt)
			if err != nil {
				e.logger.Error("encoding packet", "err", err)
				continue
			}
			e.mu.Lock()
			stream := e.tunnelStream
			e.mu.Unlock()
			if stream == nil {
				continue
			}
			if err := stream.Send(f); err != nil {
				e.logger.Error("sending frame", "err", err)
				e.handleDisconnect()
				return
			}
		}
	}
}

// pumpInbound reads frames from the server stream and dispatches packets.
func (e *Engine) pumpInbound() {
	for {
		e.mu.Lock()
		stream := e.tunnelStream
		e.mu.Unlock()
		if stream == nil {
			return
		}

		f, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			e.logger.Error("receiving frame", "err", err)
			e.handleDisconnect()
			return
		}

		switch f.FrameType {
		case aetherv1.FrameType_FRAME_TYPE_DATA:
			pkt, err := e.decoder.Decode(f)
			if err != nil {
				e.logger.Error("decoding frame", "err", err)
				continue
			}
			if pkt != nil {
				select {
				case e.inbound <- pkt:
				default:
					e.logger.Debug("inbound queue full, dropping packet")
				}
			}
		case aetherv1.FrameType_FRAME_TYPE_KEEPALIVE_ACK:
			// Keep-alive acknowledged.
		case aetherv1.FrameType_FRAME_TYPE_CLOSE:
			e.logger.Info("server sent CLOSE")
			e.handleDisconnect()
			return
		}
	}
}

// keepAliveLoop sends periodic keep-alive frames.
func (e *Engine) keepAliveLoop() {
	ticker := time.NewTicker(e.cfg.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			stream := e.tunnelStream
			e.mu.Unlock()
			if stream == nil {
				return
			}
			ka := e.encoder.EncodeKeepAlive()
			if err := stream.Send(ka); err != nil {
				e.logger.Error("sending keepalive", "err", err)
				e.handleDisconnect()
				return
			}
		}
	}
}

// handleDisconnect handles disconnection and triggers reconnect if enabled.
func (e *Engine) handleDisconnect() {
	e.mu.Lock()
	wasConnected := e.connected
	e.connected = false
	e.tunnelStream = nil
	e.mu.Unlock()

	if !wasConnected {
		return
	}

	e.logger.Warn("disconnected from server")

	if e.cfg.ReconnectEnabled {
		go e.reconnectLoop()
	}
}

// reconnectLoop attempts to reconnect with exponential backoff.
func (e *Engine) reconnectLoop() {
	backoff := e.cfg.InitialBackoff
	attempts := 0

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-time.After(backoff):
		}

		attempts++
		if e.cfg.MaxReconnect > 0 && attempts > e.cfg.MaxReconnect {
			e.logger.Error("max reconnect attempts reached")
			return
		}

		e.logger.Info("reconnecting", "attempt", attempts, "backoff", backoff)

		if err := e.Connect(); err != nil {
			e.logger.Error("reconnect failed", "err", err)
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > e.cfg.MaxBackoff {
				backoff = e.cfg.MaxBackoff
			}
			continue
		}

		e.logger.Info("reconnected successfully")
		return
	}
}

// WritePacket queues a packet to be sent through the tunnel.
func (e *Engine) WritePacket(packet []byte) {
	pkt := make([]byte, len(packet))
	copy(pkt, packet)
	select {
	case e.outbound <- pkt:
	default:
		e.logger.Debug("outbound queue full, dropping packet")
	}
}

// ReadPacket reads the next packet received from the tunnel.
func (e *Engine) ReadPacket(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case pkt := <-e.inbound:
		return pkt, nil
	}
}

// IsConnected returns the current connection state.
func (e *Engine) IsConnected() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.connected
}

// TunnelIP returns the assigned tunnel IP address.
func (e *Engine) TunnelIP() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.tunnelIP
}

// MTU returns the negotiated MTU.
func (e *Engine) MTU() uint32 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.mtu
}

// Disconnect gracefully closes the tunnel.
func (e *Engine) Disconnect() {
	e.mu.Lock()
	if e.tunnelStream != nil {
		closeFrame := e.encoder.EncodeClose()
		_ = e.tunnelStream.Send(closeFrame)
	}
	if e.conn != nil {
		e.conn.Close()
	}
	e.connected = false
	e.tunnelStream = nil
	e.mu.Unlock()

	e.cancel()
	e.logger.Info("disconnected")
}