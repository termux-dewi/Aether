package server

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aether-l3/aether/pkg/tun"
)

// TUNPacketHandler bridges gRPC tunnel streams to the server TUN device.
// It implements both PacketSink and PacketSource interfaces.
type TUNPacketHandler struct {
	device *tun.Device
	logger *slog.Logger

	// Per-client packet queues keyed by tunnel IP.
	mu      sync.Mutex
	queues  map[string]chan []byte
}

// NewTUNPacketHandler creates a handler that reads/writes to the TUN device.
func NewTUNPacketHandler(device *tun.Device, logger *slog.Logger) *TUNPacketHandler {
	return &TUNPacketHandler{
		device: device,
		logger: logger,
		queues: make(map[string]chan []byte),
	}
}

// WritePacket injects a raw IP packet into the server TUN device.
// The kernel will route it to the internet via NAT.
func (h *TUNPacketHandler) WritePacket(packet []byte) error {
	_, err := h.device.Write(packet)
	if err != nil {
		return fmt.Errorf("writing packet to TUN: %w", err)
	}
	return nil
}

// ReadPacket reads a packet from the TUN device destined for the given tunnel IP.
// This blocks until a packet is available or context is cancelled.
func (h *TUNPacketHandler) ReadPacket(ctx context.Context, tunnelIP string) ([]byte, error) {
	// Register a queue for this client if not exists.
	h.mu.Lock()
	ch, ok := h.queues[tunnelIP]
	if !ok {
		ch = make(chan []byte, 256)
		h.queues[tunnelIP] = ch
	}
	h.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case pkt := <-ch:
		return pkt, nil
	}
}

// StartReadLoop starts the goroutine that reads packets from TUN and dispatches them.
func (h *TUNPacketHandler) StartReadLoop(ctx context.Context) {
	go func() {
		buf := make([]byte, 65535)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := h.device.Read(buf)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				h.logger.Error("reading from TUN", "err", err)
				continue
			}

			if n == 0 {
				continue
			}

			// Extract destination IP to route to correct client.
			dstIP := extractDstIP(buf[:n])
			if dstIP == "" {
				continue
			}

			packet := make([]byte, n)
			copy(packet, buf[:n])

			h.mu.Lock()
			ch, ok := h.queues[dstIP]
			h.mu.Unlock()

			if ok {
				select {
				case ch <- packet:
				default:
					// Queue full, drop packet.
					h.logger.Debug("client queue full, dropping packet", "dst", dstIP)
				}
			}
		}
	}()
}

// RegisterClient adds a client queue for the given tunnel IP.
func (h *TUNPacketHandler) RegisterClient(tunnelIP string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.queues[tunnelIP]; !ok {
		h.queues[tunnelIP] = make(chan []byte, 256)
	}
}

// UnregisterClient removes a client queue.
func (h *TUNPacketHandler) UnregisterClient(tunnelIP string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.queues, tunnelIP)
}

// extractDstIP extracts the destination IP string from a raw IP packet.
func extractDstIP(packet []byte) string {
	if len(packet) < 1 {
		return ""
	}
	version := packet[0] >> 4
	switch version {
	case 4:
		if len(packet) < 20 {
			return ""
		}
		return fmt.Sprintf("%d.%d.%d.%d", packet[16], packet[17], packet[18], packet[19])
	case 6:
		if len(packet) < 40 {
			return ""
		}
		// Simplified: use net.IP for IPv6
		return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
			packet[24], packet[26], packet[28], packet[30],
			packet[32], packet[34], packet[36], packet[38])
	default:
		return ""
	}
}