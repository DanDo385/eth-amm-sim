// Package config contains chain and session configuration.
//
// SYSTEM ROLE:
// Provides the connection settings (RPC URL, chain ID) and a bridge between the
// Foundry deployment output and the Go backend. After deploy.sh runs, contract
// addresses are saved in contracts/broadcast/Deploy.s.sol/31337/run-latest.json.
// LoadAddressesFromBroadcast reads that file so the backend knows which on-chain
// contracts to interact with - no manual .env file needed.
//
// See also: config/accounts.go for the account roster and bot parameters.
package config

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// DefaultConfig returns the default configuration for Anvil.
// RPCURL override priority: ETH_AMM_SIM_RPC_URL > RPC_URL > http://localhost:8545.
func DefaultConfig() *Config {
	rpcURL := "http://localhost:8545"
	if v := strings.TrimSpace(os.Getenv("ETH_AMM_SIM_RPC_URL")); v != "" {
		rpcURL = v
	} else if v := strings.TrimSpace(os.Getenv("RPC_URL")); v != "" {
		rpcURL = v
	}
	return &Config{
		RPCURL:  rpcURL,
		ChainID: big.NewInt(31337), // Anvil default chain ID
	}
}

// BroadcastTransaction represents a transaction in Foundry's broadcast JSON
type BroadcastTransaction struct {
	ContractName    string `json:"contractName"`
	ContractAddress string `json:"contractAddress"`
}

// BroadcastFile represents the structure of Foundry's broadcast JSON
type BroadcastFile struct {
	Transactions []BroadcastTransaction `json:"transactions"`
}

// LoadAddressesFromBroadcast attempts to load contract addresses from Foundry's broadcast output
// Returns tokenAddr, ammAddr, and an error if addresses couldn't be found
func LoadAddressesFromBroadcast() (string, string, error) {
	// Find the project root (assuming we're in backend/cmd/simulator or backend/internal/config)
	// Look for contracts directory
	projectRoot := findProjectRoot()
	if projectRoot == "" {
		return "", "", fmt.Errorf("could not find project root")
	}

	broadcastDir := filepath.Join(projectRoot, "contracts", "broadcast", "Deploy.s.sol", "31337")

	// Try run-latest.json first (deploy script keeps this as the canonical latest run file)
	latestPath := filepath.Join(broadcastDir, "run-latest.json")
	if _, err := os.Stat(latestPath); os.IsNotExist(err) {
		// Find most recent run-*.json file
		entries, err := os.ReadDir(broadcastDir)
		if err != nil {
			return "", "", fmt.Errorf("could not read broadcast directory: %w", err)
		}

		var runFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "run-") && strings.HasSuffix(entry.Name(), ".json") {
				runFiles = append(runFiles, entry.Name())
			}
		}

		if len(runFiles) == 0 {
			return "", "", fmt.Errorf("no broadcast files found")
		}

		// Sort by name (which includes timestamp) - most recent last
		sort.Strings(runFiles)
		latestPath = filepath.Join(broadcastDir, runFiles[len(runFiles)-1])
	}

	// Read and parse JSON
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return "", "", fmt.Errorf("could not read broadcast file: %w", err)
	}

	var broadcast BroadcastFile
	if err := json.Unmarshal(data, &broadcast); err != nil {
		return "", "", fmt.Errorf("could not parse broadcast JSON: %w", err)
	}

	// Extract addresses
	var tokenAddr, ammAddr string
	for _, tx := range broadcast.Transactions {
		if tx.ContractName == "AppleToken" && tokenAddr == "" {
			tokenAddr = strings.ToLower(tx.ContractAddress)
		}
		if tx.ContractName == "AppleAMM" && ammAddr == "" {
			ammAddr = strings.ToLower(tx.ContractAddress)
		}
	}

	if tokenAddr == "" || ammAddr == "" {
		return "", "", fmt.Errorf("could not find contract addresses in broadcast file")
	}

	return tokenAddr, ammAddr, nil
}

// findProjectRoot walks up from current directory to find the project root
// (directory containing contracts/ and backend/ directories)
func findProjectRoot() string {
	// Start from current working directory or executable location
	startDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := startDir
	for {
		// Check if this directory contains both contracts/ and backend/
		contractsPath := filepath.Join(dir, "contracts")
		backendPath := filepath.Join(dir, "backend")

		contractsExists := false
		backendExists := false

		if info, err := os.Stat(contractsPath); err == nil && info.IsDir() {
			contractsExists = true
		}
		if info, err := os.Stat(backendPath); err == nil && info.IsDir() {
			backendExists = true
		}

		if contractsExists && backendExists {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return ""
}
