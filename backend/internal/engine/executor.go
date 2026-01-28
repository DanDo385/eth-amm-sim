// executor.go — The bridge between Go bot logic and on-chain smart contracts.
//
// SYSTEM ROLE:
// Executor is the single point through which all bot and user trades reach the
// blockchain. It ABI-encodes function calls, signs transactions with each
// account's private key, and sends them to Anvil via chain/client.go.
//
// TRADE LIFECYCLE:
//
//	Bot decides to trade → calls Executor.SwapETHForApples or SwapApplesForETH
//	→ NonceManager assigns nonce → transaction signed and sent to Anvil
//	→ Executor calculates amountOut/price from constant product formula
//	→ Trade struct emitted via callbacks → main.go records in MemoryStore
//	  and broadcasts via WebSocket → frontend Blotter and PriceChart update
//
// CONNECTIONS:
//   - Contract ABI: Manually packed (packSwapETHForApples, etc.) to match
//     contracts/src/AppleAMM.sol function signatures
//   - Callers: bots/whale.go, retail.go, meanrev.go
//   - Callbacks: main.go registers OnTrade to feed MemoryStore and WebSocket
//   - Frontend: Trade data flows through server/broadcast.go → frontend WebSocket
package engine

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"eth-amm-sim/internal/chain"
	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/contracts"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Trade represents a completed trade
type Trade struct {
	TxHash    string         `json:"txHash"`
	Trader    common.Address `json:"trader"`
	Nickname  string         `json:"nickname"`
	IsBuy     bool           `json:"isBuy"`
	AmountIn  *big.Int       `json:"amountIn"`
	AmountOut *big.Int       `json:"amountOut"`
	Price     *big.Int       `json:"price"`
	Fee       *big.Int       `json:"fee"`
	Timestamp time.Time      `json:"timestamp"`
	BlockNum  uint64         `json:"blockNum"`

	// Trade flow tracking fields (for EWMA mean reversion)
	PriceBefore        *big.Int `json:"priceBefore"`        // Price before trade (ETH per APPL, scaled by 1e18)
	PriceAfter         *big.Int `json:"priceAfter"`         // Price after trade (ETH per APPL, scaled by 1e18)
	ReservesBeforeETH  *big.Int `json:"reservesBeforeETH"`  // ETH reserve before trade
	ReservesBeforeAPPL *big.Int `json:"reservesBeforeAPPL"` // APPL reserve before trade
}

// TradeCallback is called when a trade is executed
type TradeCallback func(trade *Trade)

// Executor handles trade execution against the AMM
type Executor struct {
	client       *chain.Client
	nonceManager *chain.NonceManager
	accountMgr   *chain.AccountManager

	ammAddress   common.Address
	tokenAddress common.Address

	// Contract bindings
	ammContract   *contracts.AppleAMM
	tokenContract *contracts.AppleToken

	// Callbacks for trade events
	tradeCallbacks []TradeCallback
	callbackMu     sync.RWMutex

	// Nickname lookup
	nicknames   map[common.Address]string
	nicknamesMu sync.RWMutex
}

// NewExecutor creates a new trade executor
func NewExecutor(
	client *chain.Client,
	nonceManager *chain.NonceManager,
	ammAddress common.Address,
	tokenAddress common.Address,
) *Executor {
	// Create contract bindings
	ammContract, _ := contracts.NewAppleAMM(ammAddress, client.Client)
	tokenContract, _ := contracts.NewAppleToken(tokenAddress, client.Client)

	return &Executor{
		client:        client,
		nonceManager:  nonceManager,
		accountMgr:    chain.NewAccountManager(client.ChainID()),
		ammAddress:    ammAddress,
		tokenAddress:  tokenAddress,
		ammContract:   ammContract,
		tokenContract: tokenContract,
		nicknames:     make(map[common.Address]string),
	}
}

// SetNickname sets the nickname for an address
func (e *Executor) SetNickname(addr common.Address, nickname string) {
	e.nicknamesMu.Lock()
	e.nicknames[addr] = nickname
	e.nicknamesMu.Unlock()
}

// GetNickname gets the nickname for an address
func (e *Executor) GetNickname(addr common.Address) string {
	e.nicknamesMu.RLock()
	defer e.nicknamesMu.RUnlock()
	if nick, ok := e.nicknames[addr]; ok {
		return nick
	}
	return addr.Hex()[:10] + "..."
}

// OnTrade registers a callback for trade events
func (e *Executor) OnTrade(callback TradeCallback) {
	e.callbackMu.Lock()
	e.tradeCallbacks = append(e.tradeCallbacks, callback)
	e.callbackMu.Unlock()
}

// emitTrade calls all registered callbacks
func (e *Executor) emitTrade(trade *Trade) {
	e.callbackMu.RLock()
	callbacks := make([]TradeCallback, len(e.tradeCallbacks))
	copy(callbacks, e.tradeCallbacks)
	e.callbackMu.RUnlock()

	for _, cb := range callbacks {
		cb := cb // Capture loop variable
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Trade callback panicked: %v", r)
				}
			}()
			cb(trade)
		}()
	}
}

// SwapETHForApples swaps ETH for APPL tokens
func (e *Executor) SwapETHForApples(ctx context.Context, privateKey *ecdsa.PrivateKey, ethAmount *big.Int) (string, error) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Get nonce
	nonce, err := e.nonceManager.GetAndIncrement(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Build transaction options
	auth, err := e.accountMgr.GetTransactOptsWithValue(privateKey, ethAmount)
	if err != nil {
		return "", fmt.Errorf("failed to create tx opts: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Context = ctx

	// ABI for swapETHForApples(uint256 minApples)
	// We use 0 for minApples in simulation (no slippage protection)
	data := packSwapETHForApples(big.NewInt(0))

	tx := types.NewTransaction(
		nonce,
		e.ammAddress,
		ethAmount,
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(e.client.ChainID()), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	err = e.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send tx: %w", err)
	}

	// Calculate amount out and price using constant product formula
	// Get reserves BEFORE the trade (for trade flow tracking)
	appleReserveBefore, ethReserveBefore, err := e.GetReserves(ctx)
	if err != nil {
		// If we can't get reserves, still record trade without amounts
		trade := &Trade{
			TxHash:    signedTx.Hash().Hex(),
			Trader:    addr,
			Nickname:  e.GetNickname(addr),
			IsBuy:     true,
			AmountIn:  ethAmount,
			Timestamp: time.Now(),
		}
		e.emitTrade(trade)
		return signedTx.Hash().Hex(), nil
	}

	// Calculate price BEFORE trade: price = ethReserve / appleReserve (ETH per APPL)
	var priceBefore *big.Int
	if appleReserveBefore.Sign() > 0 {
		priceBefore = new(big.Int).Div(new(big.Int).Mul(ethReserveBefore, big.NewInt(1e18)), appleReserveBefore)
	} else {
		priceBefore = big.NewInt(0)
	}

	// Calculate fee using config constants
	fee := new(big.Int).Div(new(big.Int).Mul(ethAmount, big.NewInt(config.AMMFeeNumerator)), big.NewInt(config.AMMFeeDenominator))
	amountInAfterFee := new(big.Int).Sub(ethAmount, fee)

	// Calculate amount out: (amountInAfterFee * appleReserve) / (ethReserve + amountInAfterFee)
	numerator := new(big.Int).Mul(amountInAfterFee, appleReserveBefore)
	denominator := new(big.Int).Add(ethReserveBefore, amountInAfterFee)
	appleAmountOut := new(big.Int).Div(numerator, denominator)

	// Calculate reserves AFTER trade (for price after calculation)
	ethReserveAfter := new(big.Int).Add(ethReserveBefore, ethAmount) // Full amount including fee
	appleReserveAfter := new(big.Int).Sub(appleReserveBefore, appleAmountOut)

	// Calculate price AFTER trade
	var priceAfter *big.Int
	if appleReserveAfter.Sign() > 0 {
		priceAfter = new(big.Int).Div(new(big.Int).Mul(ethReserveAfter, big.NewInt(1e18)), appleReserveAfter)
	} else {
		priceAfter = priceBefore
	}

	// Calculate execution price: ETH spent / APPL received (scaled by 1e18)
	var price *big.Int
	if appleAmountOut.Sign() > 0 {
		// price = (ethAmount * 1e18) / appleAmountOut
		price = new(big.Int).Div(new(big.Int).Mul(ethAmount, big.NewInt(1e18)), appleAmountOut)
	} else {
		price = big.NewInt(0)
	}

	// Store price before/after and reserves before for trade flow tracking
	// These will be used by the trade callback to emit flow events

	trade := &Trade{
		TxHash:             signedTx.Hash().Hex(),
		Trader:             addr,
		Nickname:           e.GetNickname(addr),
		IsBuy:              true,
		AmountIn:           ethAmount,
		AmountOut:          appleAmountOut,
		Price:              price,
		Fee:                fee,
		Timestamp:          time.Now(),
		PriceBefore:        priceBefore,
		PriceAfter:         priceAfter,
		ReservesBeforeETH:  ethReserveBefore,
		ReservesBeforeAPPL: appleReserveBefore,
	}
	e.emitTrade(trade)

	return signedTx.Hash().Hex(), nil
}

// SwapApplesForETH swaps APPL tokens for ETH
func (e *Executor) SwapApplesForETH(ctx context.Context, privateKey *ecdsa.PrivateKey, appleAmount *big.Int) (string, error) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// First, ensure token approval if needed
	if err := e.ensureApproval(ctx, privateKey, appleAmount); err != nil {
		return "", fmt.Errorf("failed to ensure approval: %w", err)
	}

	// Get nonce
	nonce, err := e.nonceManager.GetAndIncrement(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	auth, err := e.accountMgr.GetTransactOpts(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create tx opts: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Context = ctx

	// ABI for swapApplesForETH(uint256 appleAmount, uint256 minETH)
	data := packSwapApplesForETH(appleAmount, big.NewInt(0))

	tx := types.NewTransaction(
		nonce,
		e.ammAddress,
		big.NewInt(0),
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(e.client.ChainID()), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	err = e.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send tx: %w", err)
	}

	// Calculate amount out and price using constant product formula
	// Get reserves BEFORE the trade (for trade flow tracking)
	appleReserveBefore, ethReserveBefore, err := e.GetReserves(ctx)
	if err != nil {
		// If we can't get reserves, still record trade without amounts
		trade := &Trade{
			TxHash:    signedTx.Hash().Hex(),
			Trader:    addr,
			Nickname:  e.GetNickname(addr),
			IsBuy:     false,
			AmountIn:  appleAmount,
			Timestamp: time.Now(),
		}
		e.emitTrade(trade)
		return signedTx.Hash().Hex(), nil
	}

	// Calculate price BEFORE trade: price = ethReserve / appleReserve (ETH per APPL)
	var priceBefore *big.Int
	if appleReserveBefore.Sign() > 0 {
		priceBefore = new(big.Int).Div(new(big.Int).Mul(ethReserveBefore, big.NewInt(1e18)), appleReserveBefore)
	} else {
		priceBefore = big.NewInt(0)
	}

	// Calculate fee using config constants
	fee := new(big.Int).Div(new(big.Int).Mul(appleAmount, big.NewInt(config.AMMFeeNumerator)), big.NewInt(config.AMMFeeDenominator))
	amountInAfterFee := new(big.Int).Sub(appleAmount, fee)

	// Calculate amount out: (amountInAfterFee * ethReserve) / (appleReserve + amountInAfterFee)
	numerator := new(big.Int).Mul(amountInAfterFee, ethReserveBefore)
	denominator := new(big.Int).Add(appleReserveBefore, amountInAfterFee)
	ethAmountOut := new(big.Int).Div(numerator, denominator)

	// Calculate reserves AFTER trade
	appleReserveAfter := new(big.Int).Add(appleReserveBefore, appleAmount) // Full amount including fee
	ethReserveAfter := new(big.Int).Sub(ethReserveBefore, ethAmountOut)

	// Calculate price AFTER trade
	var priceAfter *big.Int
	if appleReserveAfter.Sign() > 0 {
		priceAfter = new(big.Int).Div(new(big.Int).Mul(ethReserveAfter, big.NewInt(1e18)), appleReserveAfter)
	} else {
		priceAfter = priceBefore
	}

	// Calculate execution price: ETH received / APPL sold (scaled by 1e18)
	var price *big.Int
	if appleAmount.Sign() > 0 {
		// price = (ethAmountOut * 1e18) / appleAmount
		price = new(big.Int).Div(new(big.Int).Mul(ethAmountOut, big.NewInt(1e18)), appleAmount)
	} else {
		price = big.NewInt(0)
	}

	trade := &Trade{
		TxHash:             signedTx.Hash().Hex(),
		Trader:             addr,
		Nickname:           e.GetNickname(addr),
		IsBuy:              false,
		AmountIn:           appleAmount,
		AmountOut:          ethAmountOut,
		Price:              price,
		Fee:                fee,
		Timestamp:          time.Now(),
		PriceBefore:        priceBefore,
		PriceAfter:         priceAfter,
		ReservesBeforeETH:  ethReserveBefore,
		ReservesBeforeAPPL: appleReserveBefore,
	}
	e.emitTrade(trade)

	return signedTx.Hash().Hex(), nil
}

// ensureApproval ensures the AMM has approval to spend tokens
func (e *Executor) ensureApproval(ctx context.Context, privateKey *ecdsa.PrivateKey, amount *big.Int) error {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Check current allowance
	allowance, err := e.getAllowance(ctx, addr)
	if err != nil {
		return err
	}

	// If allowance is sufficient, no need to approve
	if allowance.Cmp(amount) >= 0 {
		return nil
	}

	// Approve max uint256
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	nonce, err := e.nonceManager.GetAndIncrement(ctx, addr)
	if err != nil {
		return err
	}

	auth, err := e.accountMgr.GetTransactOpts(privateKey)
	if err != nil {
		return err
	}
	auth.Nonce = big.NewInt(int64(nonce))

	data := packApprove(e.ammAddress, maxUint256)

	tx := types.NewTransaction(
		nonce,
		e.tokenAddress,
		big.NewInt(0),
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(e.client.ChainID()), privateKey)
	if err != nil {
		return err
	}

	return e.client.SendTransaction(ctx, signedTx)
}

// getAllowance gets the current token allowance
func (e *Executor) getAllowance(ctx context.Context, owner common.Address) (*big.Int, error) {
	data := packAllowance(owner, e.ammAddress)

	msg := ethereum.CallMsg{
		To:   &e.tokenAddress,
		Data: data,
	}

	result, err := e.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

// GetReserves returns current AMM reserves
func (e *Executor) GetReserves(ctx context.Context) (appleReserve, ethReserve *big.Int, err error) {
	result, err := e.ammContract.GetReserves(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, err
	}
	return result.AppleReserve, result.EthReserve, nil
}

// GetSpotPrice returns current spot price (ETH per APPL, scaled by 1e18)
func (e *Executor) GetSpotPrice(ctx context.Context) (*big.Int, error) {
	return e.ammContract.GetSpotPrice(&bind.CallOpts{Context: ctx})
}

// Implement bind.ContractCaller for compatibility
var _ bind.ContractCaller = (*chain.Client)(nil)

// ABI encoding helpers
func packSwapETHForApples(minApples *big.Int) []byte {
	// swapETHForApples(uint256 minApples)
	// selector: keccak256("swapETHForApples(uint256)")[:4]
	selector := []byte{0xe2, 0x6a, 0x3c, 0x8a}
	return append(selector, common.LeftPadBytes(minApples.Bytes(), 32)...)
}

func packSwapApplesForETH(appleAmount, minETH *big.Int) []byte {
	// swapApplesForETH(uint256 appleAmount, uint256 minETH)
	selector := []byte{0x48, 0xaf, 0x57, 0x21}
	data := append(selector, common.LeftPadBytes(appleAmount.Bytes(), 32)...)
	return append(data, common.LeftPadBytes(minETH.Bytes(), 32)...)
}

func packApprove(spender common.Address, amount *big.Int) []byte {
	// approve(address spender, uint256 amount)
	selector := []byte{0x09, 0x5e, 0xa7, 0xb3}
	data := append(selector, common.LeftPadBytes(spender.Bytes(), 32)...)
	return append(data, common.LeftPadBytes(amount.Bytes(), 32)...)
}

func packAllowance(owner, spender common.Address) []byte {
	// allowance(address owner, address spender)
	selector := []byte{0xdd, 0x62, 0xed, 0x3e}
	data := append(selector, common.LeftPadBytes(owner.Bytes(), 32)...)
	return append(data, common.LeftPadBytes(spender.Bytes(), 32)...)
}

// GetTotalFees returns total fees collected by the AMM
func (e *Executor) GetTotalFees(ctx context.Context) (appleFees, ethFees *big.Int, err error) {
	result, err := e.ammContract.GetTotalFees(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, err
	}
	return result.FeesApple, result.FeesETH, nil
}

// GetETHBalance returns the ETH balance of an address
func (e *Executor) GetETHBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return e.client.GetBalance(ctx, address)
}

// GetAPPLBalance returns the APPL token balance of an address
func (e *Executor) GetAPPLBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return e.tokenContract.BalanceOf(&bind.CallOpts{Context: ctx}, address)
}
