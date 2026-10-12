// Package server implements the aether-server daemon components.
package server

import (
	"sync"
	"time"
)

// Session represents an authenticated client tunnel session.
type Session struct {
	ID        []byte
	Token     []byte
	TunnelIP  string
	ClientID  string
	CreatedAt time.Time

	// LastKeepAlive tracks the last keep-alive timestamp.
	LastKeepAlive time.Time

	// Stats counters.
	BytesSent     uint64
	BytesReceived uint64
	PacketsSent   uint64
	PacketsRecv   uint64

	// Active connection counters by protocol.
	ActiveTCP  int
	ActiveUDP  int
	ActiveICMP int

	mu sync.Mutex
}

// UpdateStats updates session statistics (thread-safe).
func (s *Session) UpdateStats(bytesSent, bytesRecv uint64, pktsSent, pktsRecv uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesSent += bytesSent
	s.BytesReceived += bytesRecv
	s.PacketsSent += pktsSent
	s.PacketsRecv += pktsRecv
}

// Touch updates the last keep-alive time.
func (s *Session) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastKeepAlive = time.Now()
}

// IsExpired checks if the session has exceeded the keep-alive timeout.
func (s *Session) IsExpired(timeout time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.LastKeepAlive) > timeout
}

// SessionManager manages active tunnel sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session // keyed by hex session token
	maxConns int
	nextIP   uint32
	baseIP   uint32
}

// NewSessionManager creates a new session manager.
func NewSessionManager(maxClients int, subnetBaseIP uint32) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		maxConns: maxClients,
		nextIP:   2, // First client gets .2 (e.g., 10.8.0.2)
		baseIP:   subnetBaseIP,
	}
}

// CreateSession registers a new session. Returns error if at capacity.
func (m *SessionManager) CreateSession(token []byte, clientID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.sessions) >= m.maxConns {
		return nil, ErrMaxSessions
	}

	key := string(token)
	ip := m.baseIP + m.nextIP
	m.nextIP++

	sess := &Session{
		Token:         token,
		ClientID:      clientID,
		TunnelIP:      formatIP(ip),
		CreatedAt:     time.Now(),
		LastKeepAlive: time.Now(),
	}

	m.sessions[key] = sess
	return sess, nil
}

// GetSession retrieves a session by token.
func (m *SessionManager) GetSession(token []byte) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[string(token)]
	return sess, ok
}

// RemoveSession removes a session.
func (m *SessionManager) RemoveSession(token []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, string(token))
}

// Count returns the number of active sessions.
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// CleanupExpired removes sessions that have exceeded the timeout.
func (m *SessionManager) CleanupExpired(timeout time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	removed := 0
	for key, sess := range m.sessions {
		if sess.IsExpired(timeout) {
			delete(m.sessions, key)
			removed++
		}
	}
	return removed
}

// formatIP converts a uint32 to dotted-quad string.
func formatIP(ip uint32) string {
	return string([]byte{
		byte(ip >> 24),
		byte(ip >> 16),
		byte(ip >> 8),
		byte(ip),
	})
}