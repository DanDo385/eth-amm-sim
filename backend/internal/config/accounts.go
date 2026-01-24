// Package config contains account configuration - single source of truth
package config

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BotType identifies the trading strategy for an account
type BotType string

const (
	BotTypeLP         BotType = "lp"
	BotTypeWhale      BotType = "whale"
	BotTypeMeanRev    BotType = "meanrev"
	BotTypeMomentum   BotType = "momentum"
	BotTypeRetail     BotType = "retail"
	BotTypeLeverage   BotType = "leverage"
	BotTypeLiquidator BotType = "liquidator"
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
	LookbackBlocks int     // For meanrev/momentum: blocks to look back
	TriggerSigma   float64 // For meanrev: standard deviations to trigger
	TriggerPercent float64 // For momentum: percent move to trigger (0.03 = 3%)

	// Leverage parameters (Phase 2)
	Leverage   int      // Leverage multiplier (5, 10, 25)
	Collateral *big.Int // ETH collateral for leveraged positions
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
	"8b24eb69a6aae9d2d0e73ca675f7e3dca8e6a89df93e20d3e4049fd8a84b7b0c", // 20
	"31e52d7d8b04f0df3f7b7e4f1d3cf46e47e31f38f6cb7c1b7f5dfc5b1e7f1d5f", // 21
	"1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b", // 22
	"2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c", // 23
	"3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d", // 24
	"4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e", // 25
	"5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f", // 26
	"6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a", // 27
	"7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b", // 28
	"8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c", // 29
}

// eth is a helper to convert whole token amounts to wei (18 decimals)
func eth(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18))
}

// Accounts is the complete list of all accounts in the simulation
var Accounts []AccountConfig

// Pool configuration
var (
	PoolApples = eth(1000) // LP deposits 1000 APPLES
	PoolETH    = eth(1000) // LP deposits 1000 ETH
)

func init() {
	// Build accounts list programmatically
	Accounts = []AccountConfig{
		// ═══════════════════════════════════════════════════════════════
		// LIQUIDITY PROVIDER (Index 0)
		// ═══════════════════════════════════════════════════════════════
		{
			Index:          0,
			Nickname:       "LP",
			Type:           BotTypeLP,
			StartingApples: eth(2000), // 1000 for pool + 1000 extra
			// LP doesn't trade, so no trading params needed
		},

		// ═══════════════════════════════════════════════════════════════
		// WHALES (Index 1-3)
		// Large traders with different starting positions
		// ═══════════════════════════════════════════════════════════════
		{
			Index:          1,
			Nickname:       "Whale1",
			Type:           BotTypeWhale,
			StartingApples: eth(1000), // Already long, tends to sell
			MaxTradeSize:   eth(500),
			MaxPosition:    eth(2000),
			TradeFreqMin:   15,
			TradeFreqMax:   30,
		},
		{
			Index:          2,
			Nickname:       "Whale2",
			Type:           BotTypeWhale,
			StartingApples: eth(0), // Flat, tends to buy
			MaxTradeSize:   eth(500),
			MaxPosition:    eth(2000),
			TradeFreqMin:   15,
			TradeFreqMax:   30,
		},
		{
			Index:          3,
			Nickname:       "Whale3",
			Type:           BotTypeWhale,
			StartingApples: eth(500), // Partial position, opportunistic
			MaxTradeSize:   eth(300),
			MaxPosition:    eth(1500),
			TradeFreqMin:   15,
			TradeFreqMax:   30,
		},

		// ═══════════════════════════════════════════════════════════════
		// MEAN REVERSION (Index 4-6)
		// Fade extreme price moves, different volatility thresholds
		// Lower sigma = more sensitive, trades more frequently with smaller sizes
		// ═══════════════════════════════════════════════════════════════
		{
			Index:          4,
			Nickname:       "MeanRev1",
			Type:           BotTypeMeanRev,
			StartingApples: eth(100),
			MaxTradeSize:   eth(40), // Smaller size for more frequent 1.0 std trades
			MaxPosition:    eth(300),
			LookbackBlocks: 20,  // Medium lookback window
			TriggerSigma:   1.0, // Most sensitive - trades at 1 std deviation
		},
		{
			Index:          5,
			Nickname:       "MeanRev2",
			Type:           BotTypeMeanRev,
			StartingApples: eth(100),
			MaxTradeSize:   eth(60), // Medium size for 1.25 std trades
			MaxPosition:    eth(400),
			LookbackBlocks: 20,   // Medium lookback window
			TriggerSigma:   1.25, // Medium sensitivity
		},
		{
			Index:          6,
			Nickname:       "MeanRev3",
			Type:           BotTypeMeanRev,
			StartingApples: eth(100),
			MaxTradeSize:   eth(80), // Larger size for less frequent 1.5 std trades
			MaxPosition:    eth(500),
			LookbackBlocks: 20,  // Medium lookback window
			TriggerSigma:   1.5, // Least sensitive - trades at 1.5 std deviation
		},

		// ═══════════════════════════════════════════════════════════════
		// MOMENTUM (Index 7-9)
		// Chase trends, different sensitivities
		// ═══════════════════════════════════════════════════════════════
		{
			Index:          7,
			Nickname:       "Momentum1",
			Type:           BotTypeMomentum,
			StartingApples: eth(100),
			MaxTradeSize:   eth(50),
			MaxPosition:    eth(300),
			LookbackBlocks: 10,
			TriggerPercent: 0.03, // 3% move triggers
		},
		{
			Index:          8,
			Nickname:       "Momentum2",
			Type:           BotTypeMomentum,
			StartingApples: eth(100),
			MaxTradeSize:   eth(75),
			MaxPosition:    eth(400),
			LookbackBlocks: 20,
			TriggerPercent: 0.05, // 5% move triggers
		},
		{
			Index:          9,
			Nickname:       "Momentum3",
			Type:           BotTypeMomentum,
			StartingApples: eth(100),
			MaxTradeSize:   eth(100),
			MaxPosition:    eth(500),
			LookbackBlocks: 30,
			TriggerPercent: 0.07, // 7% move triggers
		},
	}

	// ═══════════════════════════════════════════════════════════════════
	// RETAIL (Index 10-24)
	// Generate 15 retail accounts programmatically
	// ═══════════════════════════════════════════════════════════════════
	for i := 0; i < 15; i++ {
		Accounts = append(Accounts, AccountConfig{
			Index:          10 + i,
			Nickname:       fmt.Sprintf("Retail%d", i+1),
			Type:           BotTypeRetail,
			StartingApples: eth(20),
			MaxTradeSize:   eth(10),
			MaxPosition:    eth(100),
			TradeFreqMin:   2,
			TradeFreqMax:   5,
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// LEVERAGED TRADERS (Index 25-27) - Phase 2
	// Different leverage levels, different liquidation risks
	// ═══════════════════════════════════════════════════════════════════
	Accounts = append(Accounts, []AccountConfig{
		{
			Index:          25,
			Nickname:       "Lev5x",
			Type:           BotTypeLeverage,
			StartingApples: eth(0),
			Collateral:     eth(200), // 200 ETH collateral
			Leverage:       5,        // 5x = 1000 ETH notional max
		},
		{
			Index:          26,
			Nickname:       "Lev10x",
			Type:           BotTypeLeverage,
			StartingApples: eth(0),
			Collateral:     eth(100), // 100 ETH collateral
			Leverage:       10,       // 10x = 1000 ETH notional max
		},
		{
			Index:          27,
			Nickname:       "Lev25x",
			Type:           BotTypeLeverage,
			StartingApples: eth(0),
			Collateral:     eth(40), // 40 ETH collateral
			Leverage:       25,      // 25x = 1000 ETH notional max
		},
	}...)

	// ═══════════════════════════════════════════════════════════════════
	// LIQUIDATION BOT (Index 28) - Phase 2
	// ═══════════════════════════════════════════════════════════════════
	Accounts = append(Accounts, AccountConfig{
		Index:          28,
		Nickname:       "LiqBot",
		Type:           BotTypeLiquidator,
		StartingApples: eth(0),
		// LiqBot needs ETH (from Anvil default) to execute liquidations
	})
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
