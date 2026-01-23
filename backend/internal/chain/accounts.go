// Package chain provides account management
package chain

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// AccountManager provides transaction signing capabilities
type AccountManager struct {
	chainID *big.Int
}

// NewAccountManager creates a new account manager
func NewAccountManager(chainID *big.Int) *AccountManager {
	return &AccountManager{
		chainID: chainID,
	}
}

// GetAddress derives the address from a private key
func (am *AccountManager) GetAddress(privateKey *ecdsa.PrivateKey) common.Address {
	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKeyECDSA)
}

// GetTransactOpts returns transaction options for signing
func (am *AccountManager) GetTransactOpts(privateKey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, am.chainID)
	if err != nil {
		return nil, err
	}
	
	// Default gas settings for Anvil
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = big.NewInt(1000000000) // 1 gwei
	
	return auth, nil
}

// GetTransactOptsWithValue returns transaction options with a value (for payable calls)
func (am *AccountManager) GetTransactOptsWithValue(privateKey *ecdsa.PrivateKey, value *big.Int) (*bind.TransactOpts, error) {
	auth, err := am.GetTransactOpts(privateKey)
	if err != nil {
		return nil, err
	}
	auth.Value = value
	return auth, nil
}
