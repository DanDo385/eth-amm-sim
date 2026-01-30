// Package config contains account configuration — single source of truth for all
// simulation accounts.
//
// SYSTEM ROLE:
// This file defines every account's identity, starting allocation, and trading
// parameters. It is the authoritative roster shared across the system:
//
//   - Deploy.s.sol reads these same Anvil addresses to mint tokens and redistribute ETH
//   - cmd/simulator/main.go iterates Accounts to create bot goroutines
//   - engine/executor.go uses Address()/PrivateKey() to sign transactions
//   - server/handlers.go uses Accounts to list all accounts in the REST API
//   - frontend/components/AccountMetrics.tsx displays per-account performance
//
// ACCOUNT LAYOUT:
//
//	Index 0:      LP (liquidity provider, seeds pool, does not trade)
//	Index 1:      User (manual trading from the frontend TradingPanel)
//	Index 2-16:   Retail (15 small, frequent noise traders)
//	Index 17-19:  Whales (3 large, infrequent traders)
//	Index 20-22:  MeanRev (3 EWMA-based mean reversion bots with z-score triggers)
//	Index 23-29:  Reserved (available for future use)
package config

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BotType identifies the trading strategy for an account
type BotType string

const (
	BotTypeLP      BotType = "lp"
	BotTypeWhale   BotType = "whale"
	BotTypeMeanRev BotType = "meanrev"
	BotTypeRetail  BotType = "retail"

	// UserAccountIndex is the Anvil account index for the User account (manual trading from frontend)
	//
	// The User account is at index 1, immediately after the LP account.
	// This provides a clean, logical ordering: LP(0), User(1), then trading bots.
	// Bots skip the "User" nickname, so this account is reserved for manual trades from the UI.
	UserAccountIndex = 1
)

// AccountConfig defines an account's initial state and trading parameters
type AccountConfig struct {
	// Identity
	Index    int     // Anvil account index (0-29)
	Nickname string  // Human-readable name
	Type     BotType // Bot strategy type

	// Initial allocation (set during deployment)
	StartingApples *big.Int // APPLES minted at deploy time

	// Trading parameters (used by bot at runtime)
	MaxTradeSize *big.Int // Maximum size per trade
	MaxPosition  *big.Int // Maximum total position
	TradeFreqMin int      // Minimum seconds between trades
	TradeFreqMax int      // Maximum seconds between trades

	// Strategy-specific parameters
	TriggerLevels  []float64 // For meanrev: multiple sigma levels to trade at (e.g., [0.75, 1.0, 1.25])
	HalfLifeTrades int       // For meanrev: EWMA half-life in trades (decay parameter, e.g., 50 = weight decays to 50% after 50 trades)
}

// Anvil's deterministic addresses from default mnemonic:
// "test test test test test test test test test test test junk"
var AnvilAddresses = []common.Address{
	common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), // 0
	common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"), // 1
	common.HexToAddress("0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"), // 2
	common.HexToAddress("0x90F79bf6EB2c4f870365E785982E1f101E93b906"), // 3
	common.HexToAddress("0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65"), // 4
	common.HexToAddress("0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc"), // 5
	common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9"), // 6
	common.HexToAddress("0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"), // 7
	common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f"), // 8
	common.HexToAddress("0xa0Ee7A142d267C1f36714E4a8F75612F20a79720"), // 9
	common.HexToAddress("0xBcd4042DE499D14e55001CcbB24a551F3b954096"), // 10
	common.HexToAddress("0x71bE63f3384f5fb98995898A86B02Fb2426c5788"), // 11
	common.HexToAddress("0xFABB0ac9d68B0B445fB7357272Ff202C5651694a"), // 12
	common.HexToAddress("0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec"), // 13
	common.HexToAddress("0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097"), // 14
	common.HexToAddress("0xcd3B766CCDd6AE721141F452C550Ca635964ce71"), // 15
	common.HexToAddress("0x2546BcD3c84621e976D8185a91A922aE77ECEc30"), // 16
	common.HexToAddress("0xbDA5747bFD65F08deb54cb465eB87D40e51B197E"), // 17
	common.HexToAddress("0xdD2FD4581271e230360230F9337D5c0430Bf44C0"), // 18
	common.HexToAddress("0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199"), // 19
	common.HexToAddress("0x09DB0a93B389bEF724429898f539AEB7ac2Dd55f"), // 20
	common.HexToAddress("0x02484cb50AAC86Eae85610D6f4Bf026f30f6627D"), // 21
	common.HexToAddress("0x08135Da0A343E492FA2d4282F2AE34c6c5CC1BbE"), // 22
	common.HexToAddress("0x5E661B79FE2D3F6cE70F5AAC07d8Cd9abb2743F1"), // 23
	common.HexToAddress("0x61097BA76cD906d2ba4FD106E757f7Eb455fc295"), // 24
	common.HexToAddress("0xDf37F81dAAD2b0327A0A50003740e1C935C70913"), // 25
	common.HexToAddress("0x553BC17A05702530097c3677091C5BB47a3a7931"), // 26
	common.HexToAddress("0x87BdCE72c06C21cd96219BD8521bDF1F42C78b5e"), // 27
	common.HexToAddress("0x40Fc963A729c542424cD800349a7E4Ecc4896624"), // 28
	common.HexToAddress("0x9DCCe783B6464611f38631e6C851bf441907c710"), // 29
}

// Anvil's deterministic private keys (corresponding to addresses above)
var AnvilPrivateKeys = []string{
	"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", // 0
	"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d", // 1
	"5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a", // 2
	"7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6", // 3
	"47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a", // 4
	"8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba", // 5
	"92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e", // 6
	"4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356", // 7
	"dbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97", // 8
	"2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6", // 9
	"f214f2b2cd398c806f84e317254e0f0b801d0643303237d97a22a48e01628897", // 10
	"701b615bbdfb9de65240bc28bd21bbc0d996645a3dd57e7b12bc2bdf6f192c82", // 11
	"a267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1", // 12
	"47c99abed3324a2707c28affff1267e45918ec8c3f20b8aa892e8b065d2942dd", // 13
	"c526ee95bf44d8fc405a158bb884d9d1238d99f0612e9f33d006bb0789009aaa", // 14
	"8166f546bab6da521a8369cab06c5d2b9e46670292d85c875ee9ec20e84ffb61", // 15
	"ea6c44ac03bff858b476bba40716402b03e41b8e97e276d1baec7c37d42484a0", // 16
	"689af8efa8c651a91ad287602527f3af2fe9f6501a7ac4b061667b5a93e037fd", // 17
	"de9be858da4a475276426320d5e9262ecfc3ba460bfac56360bfa6c4c28b4ee0", // 18
	"df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e", // 19
	"eaa861a9a01391ed3d587d8a5a84ca56ee277629a8b02c22093a419bf240e65d", // 20
	"c511b2aa70776d4ff1d376e8537903dae36896132c90b91d52c1dfbae267cd8b", // 21
	"224b7eb7449992aac96d631d9677f7bf5888245eef6d6eeda31e62d2f29a83e4", // 22
	"4624e0802698b9769f5bdb260a3777fbd4941ad2901f5966b854f953497eec1b", // 23
	"375ad145df13ed97f8ca8e27bb21ebf2a3819e9e0a06509a812db377e533def7", // 24
	"18743e59419b01d1d846d97ea070b5a3368a3e7f6f0242cf497e1baac6972427", // 25
	"e383b226df7c8282489889170b0f68f66af6459261f4833a781acd0804fafe7a", // 26
	"f3a6b71b94f5cd909fb2dbb287da47badaa6d8bcdc45d595e2884835d8749001", // 27
	"4e249d317253b9641e477aba8dd5d8f1f7cf5250a5acadd1229693e262720a19", // 28
	"233c86e887ac435d7f7dc64979d7758d69320906a0d340d2b6518b0fd20aa998", // 29
}

// eth is a helper to convert whole token amounts to wei (18 decimals)
func eth(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18))
}

// Accounts is the complete list of all accounts in the simulation
var Accounts []AccountConfig

// Pool configuration
// On-chain pool seeded in Deploy.s.sol with 25,000 ETH / 25,000 APPL.
// Backend previously assumed a 1M/10K pool (100 ETH/APPL) for Phase 2 metrics,
// but for the demo we want the config to exactly mirror the deployed reserves
// so LP value/price lines up between Go metrics and the Solidity AMM.
var (
	PoolApples = eth(25000) // LP deposits 25,000 APPLES
	PoolETH    = eth(25000) // LP deposits 25,000 ETH (price starts at 1.0 ETH/APPL)
)

func init() {
	// Build accounts list in logical order: LP, User, Retail, Whale, MeanRev
	Accounts = []AccountConfig{
		// ═══════════════════════════════════════════════════════════════
		// LIQUIDITY PROVIDER (Index 0)
		// ═══════════════════════════════════════════════════════════════
		{
			Index:          0,
			Nickname:       "LP",
			Type:           BotTypeLP,
			StartingApples: eth(25000), // 25,000 for pool (owns the AMM reserves)
			// LP doesn't trade, so no trading params needed
		},

		// ═══════════════════════════════════════════════════════════════
		// USER ACCOUNT (Index 1)
		// Demo account for user trading via frontend
		// ═══════════════════════════════════════════════════════════════
		{
			Index:          UserAccountIndex,
			Nickname:      "User",
			Type:          BotTypeRetail, // Mark as retail for strategy-compatibility
			StartingApples: eth(1000),    // 1,000 APPL (matches deployment script)
			// User account never runs a bot (createBots skips Nickname == "User")
			// Manual trades only via frontend TradingPanel
		},

		// ═══════════════════════════════════════════════════════════════
		// RETAIL (Index 2-16)
		// 15 small, frequent noise traders
		// Parameters:
		//   - MaxTradeSize: Maximum trade size in ETH (15 ETH)
		//   - TradeFreqMin/Max: Random delay between trades (1-4 seconds)
		//   - StartingApples: Initial APPL balance (40 APPL)
		//   - MaxPosition: Maximum total position size (100 ETH equivalent)
		// ═══════════════════════════════════════════════════════════════
	}

	// Generate 15 retail accounts (indices 2-16)
	for i := 0; i < 15; i++ {
		Accounts = append(Accounts, AccountConfig{
			Index:          2 + i,
			Nickname:       fmt.Sprintf("Retail%d", i+1),
			Type:           BotTypeRetail,
			StartingApples: eth(1000), // 1,000 APPL (matches deployment script)
			MaxTradeSize:   eth(15), // Small trade size for noise
			MaxPosition:    eth(100),
			TradeFreqMin:   1, // Very frequent (1-4 seconds)
			TradeFreqMax:   4,
		})
	}

	// ═══════════════════════════════════════════════════════════════
	// WHALES (Index 17-19)
	// 3 large, infrequent traders with random buy/sell (no directional bias)
	// Parameters:
	//   - MaxTradeSize: Maximum trade size in ETH (400-600 ETH)
	//   - TradeFreqMin/Max: Random delay between trades (45-90 seconds)
	//   - StartingApples: Initial APPL balance (5,000 APPL each)
	//   - MaxPosition: Maximum total position size (1500-2000 ETH equivalent)
	// ═══════════════════════════════════════════════════════════════
	Accounts = append(Accounts, AccountConfig{
		Index:          17,
		Nickname:       "Whale1",
		Type:           BotTypeWhale,
		StartingApples: eth(5000), // 5,000 APPL (matches deployment script)
		MaxTradeSize:   eth(600),  // Large trade size
		MaxPosition:    eth(2000),
		TradeFreqMin:   45, // Infrequent (~2-4 trades per 180s session)
		TradeFreqMax:   90,
	})
	Accounts = append(Accounts, AccountConfig{
		Index:          18,
		Nickname:       "Whale2",
		Type:           BotTypeWhale,
		StartingApples: eth(5000), // 5,000 APPL (matches deployment script)
		MaxTradeSize:   eth(600),  // Large trade size
		MaxPosition:    eth(2000),
		TradeFreqMin:   45,
		TradeFreqMax:   90,
	})
	Accounts = append(Accounts, AccountConfig{
		Index:          19,
		Nickname:       "Whale3",
		Type:           BotTypeWhale,
		StartingApples: eth(5000), // 5,000 APPL (matches deployment script)
		MaxTradeSize:   eth(400),  // Medium-large trade size
		MaxPosition:    eth(1500),
		TradeFreqMin:   45,
		TradeFreqMax:   90,
	})

	// ═══════════════════════════════════════════════════════════════
	// MEAN REVERSION (Index 20-22)
	// 3 EWMA-based mean reversion bots with different sensitivities
	// Parameters:
	//   - MaxTradeSize: Trade size in ETH (50-150 ETH, converted to APPL for sells)
	//   - TriggerLevels: Z-score thresholds (e.g., [0.75, 1.0, 1.25])
	//   - HalfLifeTrades: EWMA half-life in number of trades (25-75 trades)
	//   - StartingApples: Initial APPL balance (1,000 APPL each, matches deployment)
	//   - MaxPosition: Maximum total position size (300-500 ETH equivalent)
	//
	// MeanRev1: Fast (25 trades half-life), low thresholds → trades frequently
	// MeanRev2: Medium (50 trades half-life), moderate thresholds
	// MeanRev3: Slow (75 trades half-life), high thresholds → trades rarely
	// ═══════════════════════════════════════════════════════════════
	Accounts = append(Accounts, AccountConfig{
		Index:          20,
		Nickname:       "MeanRev1",
		Type:           BotTypeMeanRev,
		StartingApples: eth(1000), // 1,000 APPL (matches deployment script)
		MaxTradeSize:   eth(50),  // Fixed trade size
		MaxPosition:    eth(300),
		TriggerLevels:  []float64{0.75, 1.0, 1.25}, // Fast: trades at lower z-score thresholds
		HalfLifeTrades: 50,                         // Fast decay: 25 trades half-life
	})
	Accounts = append(Accounts, AccountConfig{
		Index:          21,
		Nickname:       "MeanRev2",
		Type:           BotTypeMeanRev,
		StartingApples: eth(1000), // 1,000 APPL (matches deployment script)
		MaxTradeSize:   eth(75),  // Fixed trade size
		MaxPosition:    eth(400),
		TriggerLevels:  []float64{1.0, 1.5, 2.0}, // Medium: trades at moderate thresholds
		HalfLifeTrades: 100,                       // Medium decay: 50 trades half-life
	})
	Accounts = append(Accounts, AccountConfig{
		Index:          22,
		Nickname:       "MeanRev3",
		Type:           BotTypeMeanRev,
		StartingApples: eth(1000), // 1,000 APPL (matches deployment script)
		MaxTradeSize:   eth(150), // Fixed trade size (2x MeanRev2)
		MaxPosition:    eth(500),
		TriggerLevels:  []float64{1.9, 2.4}, // Large: focuses on big moves
		HalfLifeTrades: 150,                  // Slow decay: 75 trades half-life
	})

	// Note: Indices 23-29 are reserved for future use
}

// GetAccount returns the account config for a given nickname
func GetAccount(nickname string) *AccountConfig {
	for i := range Accounts {
		if Accounts[i].Nickname == nickname {
			return &Accounts[i]
		}
	}
	return nil
}

// GetAccountByIndex returns the account config for a given Anvil index
func GetAccountByIndex(index int) *AccountConfig {
	for i := range Accounts {
		if Accounts[i].Index == index {
			return &Accounts[i]
		}
	}
	return nil
}

// GetAccountsByType returns all accounts of a given type
func GetAccountsByType(botType BotType) []AccountConfig {
	var result []AccountConfig
	for _, acc := range Accounts {
		if acc.Type == botType {
			result = append(result, acc)
		}
	}
	return result
}

// GetAddress returns the Ethereum address for an account index
func GetAddress(index int) common.Address {
	if index < 0 || index >= len(AnvilAddresses) {
		return common.Address{}
	}
	return AnvilAddresses[index]
}

// GetPrivateKey returns the private key hex string for an account index
func GetPrivateKey(index int) string {
	if index < 0 || index >= len(AnvilPrivateKeys) {
		return ""
	}
	return AnvilPrivateKeys[index]
}

// Address returns the Ethereum address for this account
func (a *AccountConfig) Address() common.Address {
	return GetAddress(a.Index)
}

// PrivateKey returns the private key hex string for this account
func (a *AccountConfig) PrivateKey() string {
	return GetPrivateKey(a.Index)
}
