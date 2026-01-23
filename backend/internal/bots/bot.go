// Package bots defines the Bot interface and common utilities
package bots

import (
	"context"
	"sync"
)

// Bot interface for all trading bots
// Each bot runs in its own goroutine and respects context cancellation
type Bot interface {
	// Run starts the bot's trading loop
	// Should block until ctx is cancelled or Stop is called
	Run(ctx context.Context)
	
	// Stop gracefully stops the bot
	Stop()
	
	// Nickname returns the bot's identifier
	Nickname() string
}

// BaseBot provides common functionality for all bots
type BaseBot struct {
	nickname string
	stopCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewBaseBot creates a new base bot
func NewBaseBot(nickname string) *BaseBot {
	return &BaseBot{
		nickname: nickname,
		stopCh:   make(chan struct{}),
	}
}

// Nickname returns the bot's identifier
func (b *BaseBot) Nickname() string {
	return b.nickname
}

// Stop signals the bot to stop
func (b *BaseBot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.running {
		close(b.stopCh)
		b.running = false
	}
}

// SetRunning marks the bot as running
func (b *BaseBot) SetRunning() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = true
}

// StopCh returns the stop channel
func (b *BaseBot) StopCh() <-chan struct{} {
	return b.stopCh
}

// IsRunning returns whether the bot is running
func (b *BaseBot) IsRunning() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running
}

// TradeDirection represents buy or sell
type TradeDirection int

const (
	Buy TradeDirection = iota
	Sell
)

func (d TradeDirection) String() string {
	if d == Buy {
		return "BUY"
	}
	return "SELL"
}
