// Package config contains configuration for the simulation
package config

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Account represents a simulation participant
type Account struct {
	Nickname   string
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
	Role       string // LP, Whale, MeanRev, Momentum, Retail, Leveraged, Liquidator
}

// Config holds all simulation configuration
type Config struct {
	// Chain settings
	RPCURL      string
	ChainID     *big.Int
	
	// Contract addresses (set after deployment)
	TokenAddress common.Address
	AMMAddress   common.Address
	
	// Accounts
	Accounts []Account
	
	// Session settings
	DefaultDuration int // seconds
}

// DefaultConfig returns the default configuration for Anvil
func DefaultConfig() *Config {
	return &Config{
		RPCURL:          "http://localhost:8545",
		ChainID:         big.NewInt(31337), // Anvil default chain ID
		DefaultDuration: 180,               // 3 minutes
		Accounts:        getAnvilAccounts(),
	}
}

// getAnvilAccounts returns the default Anvil accounts with their private keys
// These are well-known test keys - DO NOT use in production
func getAnvilAccounts() []Account {
	// Anvil default private keys (first 30)
	privateKeyHexes := []string{
		"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", // 0: LP
		"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d", // 1: Whale1
		"5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a", // 2: Whale2
		"7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6", // 3: Whale3
		"47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a", // 4: MeanRev1
		"8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba", // 5: MeanRev2
		"92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e", // 6: MeanRev3
		"4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356", // 7: Momentum1
		"dbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97", // 8: Momentum2
		"2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6", // 9: Momentum3
		"f214f2b2cd398c806f84e317254e0f0b801d0643303237d97a22a48e01628897", // 10: Retail1
		"701b615bbdfb9de65240bc28bd21bbc0d996645a3dd57e7b12bc2bdf6f192c82", // 11: Retail2
		"a267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1", // 12: Retail3
		"47c99abed3324a2707c28affff1267e45918ec8c3f20b8aa892e8b065d2942dd", // 13: Retail4
		"c526ee95bf44d8fc405a158bb884d9d1238d99f0612e9f33d006bb0789009aaa", // 14: Retail5
		"8166f546bab6da521a8369cab06c5d2b9e46670292d85c875ee9ec20e84ffb61", // 15: Retail6
		"ea6c44ac03bff858b476bba40716402b03e41b8e97e276d1baec7c37d42484a0", // 16: Retail7
		"689af8efa8c651a91ad287602527f3af2fe9f6501a7ac4b061667b5a93e037fd", // 17: Retail8
		"de9be858da4a475276426320d5e9262ecfc3ba460bfac56360bfa6c4c28b4ee0", // 18: Retail9
		"df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e", // 19: Retail10
		"8b24eb69a6aae9d2d0e73ca675f7e3dca8e6a89df93e20d3e4049fd8a84b7b0c", // 20: Retail11
		"31e52d7d8b04f0df3f7b7e4f1d3cf46e47e31f38f6cb7c1b7f5dfc5b1e7f1d5f", // 21: Retail12
		"1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b", // 22: Retail13
		"2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c", // 23: Retail14
		"3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d", // 24: Retail15
		"4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e", // 25: Lev5x
		"5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f", // 26: Lev10x
		"6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a", // 27: Lev25x
		"7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b", // 28: LiqBot
		"8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c", // 29: Reserved
	}
	
	roles := []string{
		"LP",
		"Whale", "Whale", "Whale",
		"MeanRev", "MeanRev", "MeanRev",
		"Momentum", "Momentum", "Momentum",
		"Retail", "Retail", "Retail", "Retail", "Retail",
		"Retail", "Retail", "Retail", "Retail", "Retail",
		"Retail", "Retail", "Retail", "Retail", "Retail",
		"Leveraged", "Leveraged", "Leveraged",
		"Liquidator",
		"Reserved",
	}
	
	nicknames := []string{
		"LP",
		"Whale1", "Whale2", "Whale3",
		"MeanRev1", "MeanRev2", "MeanRev3",
		"Momentum1", "Momentum2", "Momentum3",
		"Retail1", "Retail2", "Retail3", "Retail4", "Retail5",
		"Retail6", "Retail7", "Retail8", "Retail9", "Retail10",
		"Retail11", "Retail12", "Retail13", "Retail14", "Retail15",
		"Lev5x", "Lev10x", "Lev25x",
		"LiqBot",
		"Reserved",
	}
	
	accounts := make([]Account, 30)
	for i := 0; i < 30; i++ {
		privateKey, err := crypto.HexToECDSA(privateKeyHexes[i])
		if err != nil {
			// This should never happen with valid keys
			panic("invalid private key at index " + string(rune(i)))
		}
		
		publicKey := privateKey.Public()
		publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
		address := crypto.PubkeyToAddress(*publicKeyECDSA)
		
		accounts[i] = Account{
			Nickname:   nicknames[i],
			Address:    address,
			PrivateKey: privateKey,
			Role:       roles[i],
		}
	}
	
	return accounts
}

// GetAccountByNickname returns an account by nickname
func (c *Config) GetAccountByNickname(nickname string) *Account {
	for i := range c.Accounts {
		if c.Accounts[i].Nickname == nickname {
			return &c.Accounts[i]
		}
	}
	return nil
}

// GetAccountsByRole returns all accounts with a given role
func (c *Config) GetAccountsByRole(role string) []Account {
	var accounts []Account
	for _, acc := range c.Accounts {
		if acc.Role == role {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}
