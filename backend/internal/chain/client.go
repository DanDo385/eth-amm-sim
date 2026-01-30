// Package chain provides Ethereum client connectivity and nonce management.
//
// SYSTEM ROLE:
// This is the lowest layer of the backend — it speaks JSON-RPC to the Anvil
// node started by scripts/start-anvil.sh. Every on-chain read (reserves,
// prices, balances) and every transaction (swaps, approvals) flows through
// Client. NonceManager (nonce.go) ensures concurrent bot goroutines never
// collide on nonces.
//
// CONNECTIONS:
//   - Used by: engine/executor.go (all contract interactions)
//   - Depends on: Anvil running at localhost:8545 (started by scripts/start-anvil.sh)
//   - Config: RPC URL and chain ID from config/config.go
package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client wraps the go-ethereum client with helper methods
type Client struct {
	*ethclient.Client 
	rpcURL  string // the RPC URL for the chain
	chainID *big.Int // the chain ID
}

// NewClient creates a new chain client
func NewClient(rpcURL string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}
	
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}
	
	return &Client{
		Client:  client,
		rpcURL:  rpcURL,
		chainID: chainID,
	}, nil
}

// ChainID returns the chain ID
func (c *Client) ChainID() *big.Int {
	return c.chainID
}

// GetBalance returns the ETH balance of an address
func (c *Client) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return c.BalanceAt(ctx, address, nil)
}

// WaitForTx waits for a transaction to be mined and returns the receipt
func (c *Client) WaitForTx(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := c.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		// Not mined yet, wait
		return nil, fmt.Errorf("transaction not yet mined: %w", err)
	}
	return receipt, nil
}

// GetBlockNumber returns the current block number
func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	return c.BlockNumber(ctx)
}

// SetBalance sets the ETH balance of an address using Anvil's setBalance RPC method
// This is an Anvil-specific cheatcode that allows setting balances directly
func (c *Client) SetBalance(ctx context.Context, address common.Address, balance *big.Int) error {
	// Format the balance as a hex string with "0x" prefix
	balanceHex := fmt.Sprintf("0x%x", balance)
	if len(balanceHex) < 3 {
		balanceHex = "0x0"
	}

	// Use raw HTTP POST to ensure hex strings are sent correctly
	// The go-ethereum RPC client might serialize parameters in a way that converts hex to decimal
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "anvil_setBalance",
		"params":  []string{address.Hex(), balanceHex},
		"id":      1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.rpcURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var rpcResp struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	return nil
}
