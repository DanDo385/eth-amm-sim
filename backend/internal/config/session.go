// Package config contains session-level configuration
package config

import "time"

// SessionConfig holds runtime configuration for a trading session
type SessionConfig struct {
	Duration    time.Duration // Default 3 minutes
	AnvilRPCURL string        // Default http://localhost:8545
	AnvilWSURL  string        // Default ws://localhost:8545
	ServerPort  int           // Default 8080
}

// DefaultSessionConfig returns the default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		Duration:    3 * time.Minute,
		AnvilRPCURL: "http://localhost:8545",
		AnvilWSURL:  "ws://localhost:8545",
		ServerPort:  8080,
	}
}
