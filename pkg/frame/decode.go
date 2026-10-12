package frame

import (
	"encoding/binary"
	"fmt"
	"sync/atomic"

	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
)

// Decoder validates and extracts IP packets from AetherFrame messages.
type Decoder struct {
	expectedSessionID []byte
	lastSequence      atomic.Uint64
}

// NewDecoder creates a new frame decoder.
func NewDecoder(expectedSessionID []byte) *Decoder {
	return &Decoder{expectedSessionID: expectedSessionID}
}

// Decode validates a frame and returns the raw IP packet if it's a DATA frame.
// Returns nil, nil for non-data frames (keepalive, close, etc.).
func (d *Decoder) Decode(f *aetherv1.AetherFrame) ([]byte, error) {
	if f == nil {
		return nil, fmt.Errorf("nil frame")
	}

	if f.Version != ProtocolVersion {
		return nil, fmt.Errorf("unsupported frame version: %d (expected %d)", f.Version, ProtocolVersion)
	}

	switch f.FrameType {
	case aetherv1.FrameType_FRAME_TYPE_DATA:
		if len(f.PacketData) == 0 {
			return nil, fmt.Errorf("DATA frame with empty packet")
		}
		if len(f.PacketData) > MaxPacketSize {
			return nil, fmt.Errorf("DATA frame packet too large: %d", len(f.PacketData))
		}
		d.lastSequence.Store(f.Sequence)
		return f.PacketData, nil

	case aetherv1.FrameType_FRAME_TYPE_KEEPALIVE,
		aetherv1.FrameType_FRAME_TYPE_KEEPALIVE_ACK,
		aetherv1.FrameType_FRAME_TYPE_CLOSE:
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown frame type: %v", f.FrameType)
	}
}

// detectIPProtocol extracts the IP protocol number from a raw IP packet.
func detectIPProtocol(packet []byte) uint8 {
	if len(packet) < 1 {
		return 0
	}
	version := packet[0] >> 4
	switch version {
	case 4:
		if len(packet) < 20 {
			return 0
		}
		return packet[9]
	case 6:
		if len(packet) < 40 {
			return 0
		}
		return packet[6]
	default:
		return 0
	}
}

// IPPacketInfo holds parsed information about an IP packet.
type IPPacketInfo struct {
	Version   uint8
	Protocol  uint8
	SrcAddr   []byte
	DstAddr   []byte
	TotalLen  uint16
	HeaderLen uint8
}

// ParseIPPacket extracts basic information from a raw IP packet.
func ParseIPPacket(packet []byte) (*IPPacketInfo, error) {
	if len(packet) < 1 {
		return nil, fmt.Errorf("empty packet")
	}

	version := packet[0] >> 4
	info := &IPPacketInfo{Version: version}

	switch version {
	case 4:
		if len(packet) < 20 {
			return nil, fmt.Errorf("IPv4 packet too short: %d bytes", len(packet))
		}
		info.HeaderLen = (packet[0] & 0x0F) * 4
		info.TotalLen = binary.BigEndian.Uint16(packet[2:4])
		info.Protocol = packet[9]
		info.SrcAddr = packet[12:16]
		info.DstAddr = packet[16:20]

	case 6:
		if len(packet) < 40 {
			return nil, fmt.Errorf("IPv6 packet too short: %d bytes", len(packet))
		}
		info.HeaderLen = 40
		info.TotalLen = uint16(40 + binary.BigEndian.Uint16(packet[4:6]))
		info.Protocol = packet[6]
		info.SrcAddr = packet[8:24]
		info.DstAddr = packet[24:40]

	default:
		return nil, fmt.Errorf("unknown IP version: %d", version)
	}

	return info, nil
}