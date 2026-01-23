// Package config contains chain and session configuration
package config

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Config holds chain and contract configuration
type Config struct {
	// Chain settings
	RPCURL  string
	ChainID *big.Int

	// Contract addresses (set after deployment)
	TokenAddress common.Address
	AMMAddress   common.Address
}

// DefaultConfig returns the default configuration for Anvil
func DefaultConfig() *Config {
	return &Config{
		RPCURL:  "http://localhost:8545",
		ChainID: big.NewInt(31337), // Anvil default chain ID
	}
}
