// Package engine provides bot orchestration
package engine

import (
	"context"
	"log"
	"sync"
)

// Bot interface for all trading bots (defined here to avoid import cycles)
type Bot interface {
	Run(ctx context.Context)
	Stop()
	Nickname() string
}

// Orchestrator manages all bot goroutines
type Orchestrator struct {
	bots      []Bot
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.RWMutex
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		bots: make([]Bot, 0),
	}
}

// Start starts all bots with a cancellable context
func (o *Orchestrator) Start(ctx context.Context) {
	o.mu.Lock()
	
	// If already running, stop first
	if o.isRunning {
		o.mu.Unlock()
		log.Printf("Orchestrator already running, stopping first...")
		o.Stop()
		o.mu.Lock()
	}
	
	// Wait for any previous goroutines to finish
	o.mu.Unlock()
	o.wg.Wait()
	o.mu.Lock()
	
	o.isRunning = true
	sessionCtx, cancel := context.WithCancel(ctx)
	o.cancel = cancel
	
	bots := make([]Bot, len(o.bots))
	copy(bots, o.bots)
	o.mu.Unlock()

	log.Printf("Orchestrator starting %d bots", len(bots))

	for _, bot := range bots {
		o.wg.Add(1)
		go func(b Bot) {
			defer o.wg.Done()
			b.Run(sessionCtx)
		}(bot)
	}
}

// Stop stops all bot goroutines
func (o *Orchestrator) Stop() {
	o.mu.Lock()
	
	if !o.isRunning {
		o.mu.Unlock()
		return // Already stopped
	}
	
	o.isRunning = false
	
	if o.cancel != nil {
		o.cancel()
		o.cancel = nil
	}

	bots := make([]Bot, len(o.bots))
	copy(bots, o.bots)
	o.mu.Unlock()

	// Stop all bots
	for _, bot := range bots {
		bot.Stop()
	}

	// Wait for all goroutines to finish
	o.wg.Wait()
	log.Println("All bots stopped")
}

// BotCount returns the number of active bots
func (o *Orchestrator) BotCount() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.bots)
}

// GetBots returns all bots
func (o *Orchestrator) GetBots() []Bot {
	o.mu.RLock()
	defer o.mu.RUnlock()
	result := make([]Bot, len(o.bots))
	copy(result, o.bots)
	return result
}

// AddBot adds a bot to the orchestrator
func (o *Orchestrator) AddBot(bot Bot) {
	o.mu.Lock()
	o.bots = append(o.bots, bot)
	o.mu.Unlock()
	log.Printf("[Orchestrator] Added bot: %s", bot.Nickname())
}

// Clear removes all bots
func (o *Orchestrator) Clear() {
	o.mu.Lock()
	o.bots = make([]Bot, 0)
	o.mu.Unlock()
	log.Printf("[Orchestrator] Cleared all bots")
}
