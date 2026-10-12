package frame

import (
	"testing"

	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
)

// buildIPv4Packet creates a minimal IPv4 packet for testing.
func buildIPv4Packet(proto uint8, payload []byte) []byte {
	totalLen := 20 + len(payload)
	pkt := make([]byte, totalLen)
	pkt[0] = 0x45       // Version 4, IHL 5
	pkt[1] = 0          // DSCP/ECN
	pkt[2] = byte(totalLen >> 8)
	pkt[3] = byte(totalLen)
	pkt[8] = 64         // TTL
	pkt[9] = proto      // Protocol
	// src: 10.8.0.2
	pkt[12], pkt[13], pkt[14], pkt[15] = 10, 8, 0, 2
	// dst: 8.8.8.8
	pkt[16], pkt[17], pkt[18], pkt[19] = 8, 8, 8, 8
	copy(pkt[20:], payload)
	return pkt
}

func TestEncoderEncodeData(t *testing.T) {
	sessionID := []byte("test-session")
	enc := NewEncoder(sessionID)

	pkt := buildIPv4Packet(ProtoTCP, []byte("hello"))
	f, err := enc.EncodeData(pkt)
	if err != nil {
		t.Fatalf("EncodeData() error: %v", err)
	}

	if f.Version != ProtocolVersion {
		t.Fatalf("expected version %d, got %d", ProtocolVersion, f.Version)
	}
	if f.FrameType != aetherv1.FrameType_FRAME_TYPE_DATA {
		t.Fatalf("expected DATA frame type, got %v", f.FrameType)
	}
	if f.IpProtocol != ProtoTCP {
		t.Fatalf("expected proto %d, got %d", ProtoTCP, f.IpProtocol)
	}
	if len(f.PacketData) != len(pkt) {
		t.Fatalf("packet data length mismatch")
	}
}

func TestEncoderEncodeDataEmpty(t *testing.T) {
	enc := NewEncoder([]byte("s"))
	_, err := enc.EncodeData(nil)
	if err == nil {
		t.Fatal("expected error for empty packet")
	}
}

func TestDecoderDecodeData(t *testing.T) {
	sessionID := []byte("test-session")
	enc := NewEncoder(sessionID)
	dec := NewDecoder(sessionID)

	pkt := buildIPv4Packet(ProtoICMP, []byte{0x08, 0x00, 0x00, 0x00})
	f, _ := enc.EncodeData(pkt)

	decoded, err := dec.Decode(f)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if len(decoded) != len(pkt) {
		t.Fatalf("decoded packet length mismatch")
	}
}

func TestDecoderDecodeKeepAlive(t *testing.T) {
	enc := NewEncoder([]byte("s"))
	dec := NewDecoder([]byte("s"))

	ka := enc.EncodeKeepAlive()
	pkt, err := dec.Decode(ka)
	if err != nil {
		t.Fatalf("Decode(keepalive) error: %v", err)
	}
	if pkt != nil {
		t.Fatal("expected nil packet for keepalive frame")
	}
}

func TestDecoderDecodeNil(t *testing.T) {
	dec := NewDecoder([]byte("s"))
	_, err := dec.Decode(nil)
	if err == nil {
		t.Fatal("expected error for nil frame")
	}
}

func TestDetectIPProtocol(t *testing.T) {
	pkt := buildIPv4Packet(ProtoUDP, nil)
	if p := detectIPProtocol(pkt); p != ProtoUDP {
		t.Fatalf("expected %d, got %d", ProtoUDP, p)
	}

	if p := detectIPProtocol(nil); p != 0 {
		t.Fatal("expected 0 for nil packet")
	}
	if p := detectIPProtocol([]byte{0x00}); p != 0 {
		t.Fatal("expected 0 for unknown version")
	}
}

func TestParseIPPacket(t *testing.T) {
	pkt := buildIPv4Packet(ProtoTCP, []byte("data"))
	info, err := ParseIPPacket(pkt)
	if err != nil {
		t.Fatalf("ParseIPPacket() error: %v", err)
	}
	if info.Version != 4 {
		t.Fatalf("expected IPv4, got %d", info.Version)
	}
	if info.Protocol != ProtoTCP {
		t.Fatalf("expected proto %d, got %d", ProtoTCP, info.Protocol)
	}
	if info.HeaderLen != 20 {
		t.Fatalf("expected header len 20, got %d", info.HeaderLen)
	}
}

func TestPacketPool(t *testing.T) {
	pool := NewPacketPool(1500)
	buf := pool.Get()
	if len(*buf) != 1500 {
		t.Fatalf("expected buffer size 1500, got %d", len(*buf))
	}
	pool.Put(buf)
}