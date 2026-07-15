// orchestrator.go - Manages the lifecycle of all bot goroutines.
//
// When a session starts (via frontend SessionControls → POST /session/start),
// Start() launches each bot's Run() method in a separate goroutine using errgroup.
// When the session ends (timeout or user stop), Stop() cancels the shared context,
// causing all bots to exit their main loops.
//
// CONNECTIONS:
//  - Created in: cmd/simulator/main.go
//  - Controlled by: engine/session.go (Start/Stop lifecycle)
//  - Bots registered via: AddBot() called from main.go:createBots()
//  - Frontend trigger: SessionControls component → useSession hook → REST API
package engine

import (
	"context"
	"log"
	"sync"

	"golang.org/x/sync/errgroup"
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
	eg        *errgroup.Group
	egCtx     context.Context
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
	if o.eg != nil {
		o.mu.Unlock()
		o.eg.Wait() // Wait for previous errgroup to finish
		o.mu.Lock()
	}
	
	o.isRunning = true
	sessionCtx, cancel := context.WithCancel(ctx)
	o.cancel = cancel
	
	// Create new errgroup with session context
	o.eg, o.egCtx = errgroup.WithContext(sessionCtx)
	
	bots := make([]Bot, len(o.bots))
	copy(bots, o.bots)
	o.mu.Unlock()

	log.Printf("Orchestrator starting %d bots", len(bots))

	for _, bot := range bots {
		bot := bot // Capture loop variable
		o.eg.Go(func() error {
			// Bot.Run doesn't return an error, but we can detect panics
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[%s] Bot panicked: %v", bot.Nickname(), r)
				}
			}()
			bot.Run(o.egCtx)
			return nil // Bots don't return errors, they run until context cancelled
		})
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
	eg := o.eg
	o.mu.Unlock()

	// Stop all bots
	for _, bot := range bots {
		bot.Stop()
	}

	// Wait for all goroutines to finish
	if eg != nil {
		if err := eg.Wait(); err != nil {
			log.Printf("Error waiting for bots: %v", err)
		}
	}
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