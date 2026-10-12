// Package frame implements the Aether L3 packet framing protocol.
package frame

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
)

const (
	ProtocolVersion = 1
	MaxPacketSize   = 65535
	DefaultMTU      = 1500
)

// IP protocol numbers.
const (
	ProtoICMP   = 1
	ProtoTCP    = 6
	ProtoUDP    = 17
	ProtoICMPv6 = 58
)

// Encoder encapsulates raw IP packets into AetherFrame messages.
type Encoder struct {
	sessionID []byte
	sequence  atomic.Uint64
}

// NewEncoder creates a new frame encoder with the given session ID.
func NewEncoder(sessionID []byte) *Encoder {
	return &Encoder{sessionID: sessionID}
}

// EncodeData wraps a raw IP packet into an AetherFrame DATA frame.
func (e *Encoder) EncodeData(packet []byte) (*aetherv1.AetherFrame, error) {
	if len(packet) == 0 {
		return nil, fmt.Errorf("empty packet")
	}
	if len(packet) > MaxPacketSize {
		return nil, fmt.Errorf("packet too large: %d bytes (max %d)", len(packet), MaxPacketSize)
	}

	proto := detectIPProtocol(packet)

	f := &aetherv1.AetherFrame{
		Version:     ProtocolVersion,
		SessionId:   e.sessionID,
		Sequence:    e.sequence.Add(1),
		PacketData:  make([]byte, len(packet)),
		IpProtocol:  uint32(proto),
		TimestampNs: time.Now().UnixNano(),
		FrameType:   aetherv1.FrameType_FRAME_TYPE_DATA,
	}
	copy(f.PacketData, packet)
	return f, nil
}

// EncodeKeepAlive creates a keep-alive frame.
func (e *Encoder) EncodeKeepAlive() *aetherv1.AetherFrame {
	return &aetherv1.AetherFrame{
		Version:     ProtocolVersion,
		SessionId:   e.sessionID,
		Sequence:    e.sequence.Add(1),
		TimestampNs: time.Now().UnixNano(),
		FrameType:   aetherv1.FrameType_FRAME_TYPE_KEEPALIVE,
	}
}

// EncodeKeepAliveAck creates a keep-alive acknowledgment frame.
func (e *Encoder) EncodeKeepAliveAck() *aetherv1.AetherFrame {
	return &aetherv1.AetherFrame{
		Version:     ProtocolVersion,
		SessionId:   e.sessionID,
		Sequence:    e.sequence.Add(1),
		TimestampNs: time.Now().UnixNano(),
		FrameType:   aetherv1.FrameType_FRAME_TYPE_KEEPALIVE_ACK,
	}
}

// EncodeClose creates a session close frame.
func (e *Encoder) EncodeClose() *aetherv1.AetherFrame {
	return &aetherv1.AetherFrame{
		Version:     ProtocolVersion,
		SessionId:   e.sessionID,
		Sequence:    e.sequence.Add(1),
		TimestampNs: time.Now().UnixNano(),
		FrameType:   aetherv1.FrameType_FRAME_TYPE_CLOSE,
	}
}

// PacketPool provides a pool of byte slices for packet buffers.
type PacketPool struct {
	pool sync.Pool
}

// NewPacketPool creates a new packet buffer pool.
func NewPacketPool(mtu int) *PacketPool {
	return &PacketPool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, mtu)
				return &buf
			},
		},
	}
}

// Get retrieves a buffer from the pool.
func (p *PacketPool) Get() *[]byte {
	return p.pool.Get().(*[]byte)
}

// Put returns a buffer to the pool.
func (p *PacketPool) Put(buf *[]byte) {
	p.pool.Put(buf)
}