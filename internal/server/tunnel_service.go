package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"

	"github.com/aether-l3/aether/pkg/auth"
	aetherv1 "github.com/aether-l3/aether/proto/aether/v1"
)

// PacketSink is the interface for handling decoded IP packets from clients.
type PacketSink interface {
	// WritePacket injects a raw IP packet into the server TUN/netstack.
	WritePacket(packet []byte) error
}

// PacketSource is the interface for reading IP packets to send to a client.
type PacketSource interface {
	// ReadPacket reads the next IP packet destined for the given tunnel IP.
	ReadPacket(ctx context.Context, tunnelIP string) ([]byte, error)
}

// TunnelServiceImpl implements the gRPC TunnelService.
type TunnelServiceImpl struct {
	aetherv1.UnimplementedTunnelServiceServer

	authenticator *auth.Authenticator
	sessions      *SessionManager
	packetSink    PacketSink
	packetSource  PacketSource
	logger        *slog.Logger
}

// NewTunnelService creates a new TunnelService gRPC implementation.
func NewTunnelService(
	authenticator *auth.Authenticator,
	sessions *SessionManager,
	sink PacketSink,
	source PacketSource,
	logger *slog.Logger,
) *TunnelServiceImpl {
	return &TunnelServiceImpl{
		authenticator: authenticator,
		sessions:      sessions,
		packetSink:    sink,
		packetSource:  source,
		logger:        logger,
	}
}

// Handshake authenticates the client and creates a session.
func (s *TunnelServiceImpl) Handshake(ctx context.Context, req *aetherv1.HandshakeRequest) (*aetherv1.HandshakeResponse, error) {
	if err := s.authenticator.VerifySimpleToken(req.AuthToken); err != nil {
		s.logger.Warn("handshake auth failed", "client", req.ClientId, "err", err)
		return &aetherv1.HandshakeResponse{
			Accepted: false,
			Reason:   "authentication failed",
		}, nil
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	sess, err := s.sessions.CreateSession(token, req.ClientId)
	if err != nil {
		return &aetherv1.HandshakeResponse{
			Accepted: false,
			Reason:   err.Error(),
		}, nil
	}

	mtu := req.RequestedMtu
	if mtu == 0 {
		mtu = 1500
	}

	s.logger.Info("client connected",
		"client", req.ClientId,
		"tunnel_ip", sess.TunnelIP,
		"token", hex.EncodeToString(token[:8]),
	)

	return &aetherv1.HandshakeResponse{
		Accepted:         true,
		SessionToken:     token,
		NegotiatedMtu:    mtu,
		TunnelIp:         sess.TunnelIP,
		TunnelPrefixLen:  24,
	}, nil
}

// Tunnel handles the bidirectional packet stream.
func (s *TunnelServiceImpl) Tunnel(stream aetherv1.TunnelService_TunnelServer) error {
	// Read the first frame to identify the session.
	firstFrame, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receiving first frame: %w", err)
	}

	sess, ok := s.sessions.GetSession(firstFrame.SessionId)
	if !ok {
		return ErrSessionNotFound
	}

	ctx := stream.Context()
	errCh := make(chan error, 2)

	// Goroutine: client -> server (read frames, inject into TUN).
	go func() {
		errCh <- s.readFromClient(stream, sess)
	}()

	// Goroutine: server -> client (read from TUN/netstack, send frames).
	go func() {
		errCh <- s.writeToClient(ctx, stream, sess)
	}()

	// Wait for either direction to finish.
	err = <-errCh

	s.logger.Info("client disconnected",
		"client", sess.ClientID,
		"tunnel_ip", sess.TunnelIP,
	)

	return err
}

// readFromClient reads frames from the client stream and injects packets.
func (s *TunnelServiceImpl) readFromClient(stream aetherv1.TunnelService_TunnelServer, sess *Session) error {
	for {
		frame, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("receiving frame: %w", err)
		}

		switch frame.FrameType {
		case aetherv1.FrameType_FRAME_TYPE_DATA:
			if len(frame.PacketData) > 0 {
				if err := s.packetSink.WritePacket(frame.PacketData); err != nil {
					s.logger.Error("injecting packet", "err", err)
				}
				sess.UpdateStats(uint64(len(frame.PacketData)), 0, 1, 0)
			}

		case aetherv1.FrameType_FRAME_TYPE_KEEPALIVE:
			sess.Touch()
			// Send keep-alive ACK back.
			ack := &aetherv1.AetherFrame{
				Version:     1,
				SessionId:   sess.Token,
				FrameType:   aetherv1.FrameType_FRAME_TYPE_KEEPALIVE_ACK,
				TimestampNs: frame.TimestampNs,
			}
			if err := stream.Send(ack); err != nil {
				return fmt.Errorf("sending keepalive ack: %w", err)
			}

		case aetherv1.FrameType_FRAME_TYPE_CLOSE:
			return nil
		}
	}
}

// writeToClient reads packets from the server netstack and sends to client.
func (s *TunnelServiceImpl) writeToClient(ctx context.Context, stream aetherv1.TunnelService_TunnelServer, sess *Session) error {
	for {
		packet, err := s.packetSource.ReadPacket(ctx, sess.TunnelIP)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("reading packet: %w", err)
		}

		frame := &aetherv1.AetherFrame{
			Version:     1,
			SessionId:   sess.Token,
			PacketData:  packet,
			FrameType:   aetherv1.FrameType_FRAME_TYPE_DATA,
		}

		if err := stream.Send(frame); err != nil {
			return fmt.Errorf("sending frame: %w", err)
		}

		sess.UpdateStats(0, uint64(len(packet)), 0, 1)
	}
}

// GetStats streams session statistics.
func (s *TunnelServiceImpl) GetStats(req *aetherv1.StatsRequest, stream aetherv1.TunnelService_GetStatsServer) error {
	sess, ok := s.sessions.GetSession(req.SessionToken)
	if !ok {
		return ErrSessionNotFound
	}

	update := &aetherv1.StatsUpdate{
		BytesSent:      sess.BytesSent,
		BytesReceived:  sess.BytesReceived,
		PacketsSent:    sess.PacketsSent,
		PacketsReceived: sess.PacketsRecv,
		ActiveTcpConns: uint32(sess.ActiveTCP),
		ActiveUdpConns: uint32(sess.ActiveUDP),
		ActiveIcmpConns: uint32(sess.ActiveICMP),
	}

	return stream.Send(update)
}