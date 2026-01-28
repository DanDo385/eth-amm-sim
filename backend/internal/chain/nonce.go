// nonce.go — Thread-safe nonce tracking for concurrent bot transactions.
//
// Each Ethereum transaction needs a sequential nonce per sender address.
// Because multiple bot goroutines submit transactions concurrently,
// NonceManager serializes nonce allocation with a mutex. Each call to
// GetAndIncrement atomically returns the next nonce and advances the counter.
//
// Used by: engine/executor.go (SwapETHForApples, SwapApplesForETH,
// ensureApproval, OpenLeveragedPosition, LiquidatePosition)
package chain

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// NonceManager manages nonces for concurrent transaction submission
// Critical for multi-bot simulation where multiple goroutines submit txs
type NonceManager struct {
	mu      sync.Mutex
	client  *Client
	nonces  map[common.Address]uint64 // tracks nonces for each address
	pending map[common.Address]uint64 // tracks pending nonces
}

// NewNonceManager creates a new nonce manager
func NewNonceManager(client *Client) *NonceManager {
	return &NonceManager{
		client:  client,
		nonces:  make(map[common.Address]uint64), // tracks nonces for each address
		pending: make(map[common.Address]uint64), // tracks pending nonces for each address
	}
}

// GetAndIncrement returns the next nonce for an address and increments it
// This is thread-safe and prevents nonce collisions between goroutines
func (nm *NonceManager) GetAndIncrement(ctx context.Context, addr common.Address) (uint64, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	// Initialize nonce from chain if not cached
	if _, exists := nm.nonces[addr]; !exists {
		pendingNonce, err := nm.client.PendingNonceAt(ctx, addr)
		if err != nil {
			return 0, err
		}
		nm.nonces[addr] = pendingNonce
	}
	
	nonce := nm.nonces[addr]
	nm.nonces[addr]++
	
	return nonce, nil
}

// PeekNonce returns the next nonce without incrementing
func (nm *NonceManager) PeekNonce(ctx context.Context, addr common.Address) (uint64, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	if _, exists := nm.nonces[addr]; !exists {
		pendingNonce, err := nm.client.PendingNonceAt(ctx, addr)
		if err != nil {
			return 0, err
		}
		nm.nonces[addr] = pendingNonce
	}
	
	return nm.nonces[addr], nil
}

// Reset resets the nonce for an address (useful after failed tx)
func (nm *NonceManager) Reset(ctx context.Context, addr common.Address) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	pendingNonce, err := nm.client.PendingNonceAt(ctx, addr)
	if err != nil {
		return err
	}
	
	nm.nonces[addr] = pendingNonce
	return nil
}

// ResetAll resets all cached nonces
func (nm *NonceManager) ResetAll(ctx context.Context) error {
	nm.mu.Lock()
	addresses := make([]common.Address, 0, len(nm.nonces))
	for addr := range nm.nonces {
		addresses = append(addresses, addr)
	}
	nm.nonces = make(map[common.Address]uint64)
	nm.mu.Unlock()
	
	// Re-fetch nonces for all addresses
	for _, addr := range addresses {
		if _, err := nm.GetAndIncrement(ctx, addr); err != nil {
			return err
		}
	}
	
	return nil
}
