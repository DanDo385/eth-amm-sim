// Package chain provides Ethereum client connectivity
package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
