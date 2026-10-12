package server

import "errors"

var (
	// ErrMaxSessions is returned when the server has reached max client capacity.
	ErrMaxSessions = errors.New("maximum number of concurrent sessions reached")

	// ErrAuthFailed is returned when client authentication fails.
	ErrAuthFailed = errors.New("authentication failed")

	// ErrSessionNotFound is returned when a session token is invalid.
	ErrSessionNotFound = errors.New("session not found")

	// ErrServerShutdown is returned when the server is shutting down.
	ErrServerShutdown = errors.New("server is shutting down")
)