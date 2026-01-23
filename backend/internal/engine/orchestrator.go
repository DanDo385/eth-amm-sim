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

// Orchestrator manages multiple bots
type Orchestrator struct {
	mu   sync.RWMutex
	bots []Bot
	wg   sync.WaitGroup
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		bots: make([]Bot, 0),
	}
}

// AddBot adds a bot to the orchestrator
func (o *Orchestrator) AddBot(bot Bot) {
	o.mu.Lock()
	o.bots = append(o.bots, bot)
	o.mu.Unlock()
	log.Printf("[Orchestrator] Added bot: %s", bot.Nickname())
}

// StartAll starts all bots
func (o *Orchestrator) StartAll(ctx context.Context) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	log.Printf("[Orchestrator] Starting %d bots", len(o.bots))
	
	for _, bot := range o.bots {
		o.wg.Add(1)
		go func(b Bot) {
			defer o.wg.Done()
			b.Run(ctx)
		}(bot)
	}
}

// StopAll stops all bots
func (o *Orchestrator) StopAll() {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	log.Printf("[Orchestrator] Stopping all bots")
	
	for _, bot := range o.bots {
		bot.Stop()
	}
	
	// Wait for all bots to complete
	o.wg.Wait()
	
	log.Printf("[Orchestrator] All bots stopped")
}

// GetBots returns all bots
func (o *Orchestrator) GetBots() []Bot {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	result := make([]Bot, len(o.bots))
	copy(result, o.bots)
	return result
}

// GetBotByNickname returns a bot by nickname
func (o *Orchestrator) GetBotByNickname(nickname string) Bot {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	for _, bot := range o.bots {
		if bot.Nickname() == nickname {
			return bot
		}
	}
	return nil
}

// Clear removes all bots
func (o *Orchestrator) Clear() {
	o.mu.Lock()
	o.bots = make([]Bot, 0)
	o.mu.Unlock()
	log.Printf("[Orchestrator] Cleared all bots")
}
