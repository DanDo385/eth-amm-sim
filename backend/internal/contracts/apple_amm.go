// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// AppleAMMPosition is an auto generated low-level Go binding around an user-defined struct.
type AppleAMMPosition struct {
	Trader     common.Address
	IsLong     bool
	Collateral *big.Int
	Size       *big.Int
	EntryPrice *big.Int
	Leverage   *big.Int
	LoanAmount *big.Int
	Timestamp  *big.Int
}

// AppleAMMMetaData contains all meta data concerning the AppleAMM contract.
var AppleAMMMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_appleToken\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"receive\",\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"FEE_DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"FEE_NUMERATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"INITIAL_MARGIN_BPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"LIQUIDATION_FEE_BPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"LIQUIDATION_THRESHOLD_BPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAINTENANCE_MARGIN_BPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"addLiquidity\",\"inputs\":[{\"name\":\"appleAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"lpTokens\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"appleReserve\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"appleToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"closePosition\",\"inputs\":[],\"outputs\":[{\"name\":\"ethOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"ethReserve\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAmountOut\",\"inputs\":[{\"name\":\"amountIn\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"reserveIn\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"reserveOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"amountOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getLPBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPosition\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"pos\",\"type\":\"tuple\",\"internalType\":\"structAppleAMM.Position\",\"components\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isLong\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"collateral\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"size\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"entryPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"leverage\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"loanAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPositionValue\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPositionValue\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"currentPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReserves\",\"inputs\":[],\"outputs\":[{\"name\":\"_appleReserve\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_ethReserve\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSpotPrice\",\"inputs\":[],\"outputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTotalFees\",\"inputs\":[],\"outputs\":[{\"name\":\"feesApple\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feesETH\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isLiquidatable\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"liquidatable\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"marginRatio\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"liquidate\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"success\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"lpBalances\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"openLeveragedPosition\",\"inputs\":[{\"name\":\"leverage\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minApples\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"applesOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"positions\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isLong\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"collateral\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"size\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"entryPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"leverage\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"loanAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"removeLiquidity\",\"inputs\":[{\"name\":\"lpTokenAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"appleOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"ethOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"swapApplesForETH\",\"inputs\":[{\"name\":\"appleAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minETH\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"ethOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"swapETHForApples\",\"inputs\":[{\"name\":\"minApples\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"applesOut\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"totalFeesApple\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalFeesETH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalLPTokens\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalLentETH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"LogLiquidityAdded\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"appleAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"ethAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"lpTokens\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"LogLiquidityRemoved\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"appleAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"ethAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"lpTokens\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"LogSwap\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"isBuy\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"},{\"name\":\"amountIn\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"amountOut\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"fee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PositionClosed\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"collateralReturned\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"pnl\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PositionLiquidated\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"liquidator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"collateral\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"liquidationFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"remainingCollateral\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PositionOpened\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"isLong\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"},{\"name\":\"collateral\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"size\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"entryPrice\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"leverage\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AmountMustBePositive\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientAPPLLiquidity\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientBalance\",\"inputs\":[{\"name\":\"balance\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"needed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InsufficientETHLiquidity\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientInitialMargin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientInputAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientLPTokens\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientLiquidity\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientOutput\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidLeverage\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTokenAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoPosition\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyLongPositionsSupported\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PoolEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PositionAlreadyExists\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PositionNotLiquidatable\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SlippageExceeded\",\"inputs\":[]}]",
}

// AppleAMMABI is the input ABI used to generate the binding from.
// Deprecated: Use AppleAMMMetaData.ABI instead.
var AppleAMMABI = AppleAMMMetaData.ABI

// AppleAMM is an auto generated Go binding around an Ethereum contract.
type AppleAMM struct {
	AppleAMMCaller     // Read-only binding to the contract
	AppleAMMTransactor // Write-only binding to the contract
	AppleAMMFilterer   // Log filterer for contract events
}

// AppleAMMCaller is an auto generated read-only Go binding around an Ethereum contract.
type AppleAMMCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppleAMMTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AppleAMMTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppleAMMFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AppleAMMFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppleAMMSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AppleAMMSession struct {
	Contract     *AppleAMM         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AppleAMMCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AppleAMMCallerSession struct {
	Contract *AppleAMMCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// AppleAMMTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AppleAMMTransactorSession struct {
	Contract     *AppleAMMTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// AppleAMMRaw is an auto generated low-level Go binding around an Ethereum contract.
type AppleAMMRaw struct {
	Contract *AppleAMM // Generic contract binding to access the raw methods on
}

// AppleAMMCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AppleAMMCallerRaw struct {
	Contract *AppleAMMCaller // Generic read-only contract binding to access the raw methods on
}

// AppleAMMTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AppleAMMTransactorRaw struct {
	Contract *AppleAMMTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAppleAMM creates a new instance of AppleAMM, bound to a specific deployed contract.
func NewAppleAMM(address common.Address, backend bind.ContractBackend) (*AppleAMM, error) {
	contract, err := bindAppleAMM(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AppleAMM{AppleAMMCaller: AppleAMMCaller{contract: contract}, AppleAMMTransactor: AppleAMMTransactor{contract: contract}, AppleAMMFilterer: AppleAMMFilterer{contract: contract}}, nil
}

// NewAppleAMMCaller creates a new read-only instance of AppleAMM, bound to a specific deployed contract.
func NewAppleAMMCaller(address common.Address, caller bind.ContractCaller) (*AppleAMMCaller, error) {
	contract, err := bindAppleAMM(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AppleAMMCaller{contract: contract}, nil
}

// NewAppleAMMTransactor creates a new write-only instance of AppleAMM, bound to a specific deployed contract.
func NewAppleAMMTransactor(address common.Address, transactor bind.ContractTransactor) (*AppleAMMTransactor, error) {
	contract, err := bindAppleAMM(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AppleAMMTransactor{contract: contract}, nil
}

// NewAppleAMMFilterer creates a new log filterer instance of AppleAMM, bound to a specific deployed contract.
func NewAppleAMMFilterer(address common.Address, filterer bind.ContractFilterer) (*AppleAMMFilterer, error) {
	contract, err := bindAppleAMM(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AppleAMMFilterer{contract: contract}, nil
}

// bindAppleAMM binds a generic wrapper to an already deployed contract.
func bindAppleAMM(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AppleAMMMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AppleAMM *AppleAMMRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AppleAMM.Contract.AppleAMMCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AppleAMM *AppleAMMRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppleAMM.Contract.AppleAMMTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AppleAMM *AppleAMMRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AppleAMM.Contract.AppleAMMTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AppleAMM *AppleAMMCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AppleAMM.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AppleAMM *AppleAMMTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppleAMM.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AppleAMM *AppleAMMTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AppleAMM.Contract.contract.Transact(opts, method, params...)
}

// FEEDENOMINATOR is a free data retrieval call binding the contract method 0xd73792a9.
//
// Solidity: function FEE_DENOMINATOR() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) FEEDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "FEE_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FEEDENOMINATOR is a free data retrieval call binding the contract method 0xd73792a9.
//
// Solidity: function FEE_DENOMINATOR() view returns(uint256)
func (_AppleAMM *AppleAMMSession) FEEDENOMINATOR() (*big.Int, error) {
	return _AppleAMM.Contract.FEEDENOMINATOR(&_AppleAMM.CallOpts)
}

// FEEDENOMINATOR is a free data retrieval call binding the contract method 0xd73792a9.
//
// Solidity: function FEE_DENOMINATOR() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) FEEDENOMINATOR() (*big.Int, error) {
	return _AppleAMM.Contract.FEEDENOMINATOR(&_AppleAMM.CallOpts)
}

// FEENUMERATOR is a free data retrieval call binding the contract method 0x41cd47bf.
//
// Solidity: function FEE_NUMERATOR() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) FEENUMERATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "FEE_NUMERATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FEENUMERATOR is a free data retrieval call binding the contract method 0x41cd47bf.
//
// Solidity: function FEE_NUMERATOR() view returns(uint256)
func (_AppleAMM *AppleAMMSession) FEENUMERATOR() (*big.Int, error) {
	return _AppleAMM.Contract.FEENUMERATOR(&_AppleAMM.CallOpts)
}

// FEENUMERATOR is a free data retrieval call binding the contract method 0x41cd47bf.
//
// Solidity: function FEE_NUMERATOR() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) FEENUMERATOR() (*big.Int, error) {
	return _AppleAMM.Contract.FEENUMERATOR(&_AppleAMM.CallOpts)
}

// INITIALMARGINBPS is a free data retrieval call binding the contract method 0xa7bd8e13.
//
// Solidity: function INITIAL_MARGIN_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) INITIALMARGINBPS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "INITIAL_MARGIN_BPS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// INITIALMARGINBPS is a free data retrieval call binding the contract method 0xa7bd8e13.
//
// Solidity: function INITIAL_MARGIN_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMSession) INITIALMARGINBPS() (*big.Int, error) {
	return _AppleAMM.Contract.INITIALMARGINBPS(&_AppleAMM.CallOpts)
}

// INITIALMARGINBPS is a free data retrieval call binding the contract method 0xa7bd8e13.
//
// Solidity: function INITIAL_MARGIN_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) INITIALMARGINBPS() (*big.Int, error) {
	return _AppleAMM.Contract.INITIALMARGINBPS(&_AppleAMM.CallOpts)
}

// LIQUIDATIONFEEBPS is a free data retrieval call binding the contract method 0x518e8dfc.
//
// Solidity: function LIQUIDATION_FEE_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) LIQUIDATIONFEEBPS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "LIQUIDATION_FEE_BPS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LIQUIDATIONFEEBPS is a free data retrieval call binding the contract method 0x518e8dfc.
//
// Solidity: function LIQUIDATION_FEE_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMSession) LIQUIDATIONFEEBPS() (*big.Int, error) {
	return _AppleAMM.Contract.LIQUIDATIONFEEBPS(&_AppleAMM.CallOpts)
}

// LIQUIDATIONFEEBPS is a free data retrieval call binding the contract method 0x518e8dfc.
//
// Solidity: function LIQUIDATION_FEE_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) LIQUIDATIONFEEBPS() (*big.Int, error) {
	return _AppleAMM.Contract.LIQUIDATIONFEEBPS(&_AppleAMM.CallOpts)
}

// LIQUIDATIONTHRESHOLDBPS is a free data retrieval call binding the contract method 0x0c6d5a58.
//
// Solidity: function LIQUIDATION_THRESHOLD_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) LIQUIDATIONTHRESHOLDBPS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "LIQUIDATION_THRESHOLD_BPS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LIQUIDATIONTHRESHOLDBPS is a free data retrieval call binding the contract method 0x0c6d5a58.
//
// Solidity: function LIQUIDATION_THRESHOLD_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMSession) LIQUIDATIONTHRESHOLDBPS() (*big.Int, error) {
	return _AppleAMM.Contract.LIQUIDATIONTHRESHOLDBPS(&_AppleAMM.CallOpts)
}

// LIQUIDATIONTHRESHOLDBPS is a free data retrieval call binding the contract method 0x0c6d5a58.
//
// Solidity: function LIQUIDATION_THRESHOLD_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) LIQUIDATIONTHRESHOLDBPS() (*big.Int, error) {
	return _AppleAMM.Contract.LIQUIDATIONTHRESHOLDBPS(&_AppleAMM.CallOpts)
}

// MAINTENANCEMARGINBPS is a free data retrieval call binding the contract method 0x5838bff7.
//
// Solidity: function MAINTENANCE_MARGIN_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) MAINTENANCEMARGINBPS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "MAINTENANCE_MARGIN_BPS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAINTENANCEMARGINBPS is a free data retrieval call binding the contract method 0x5838bff7.
//
// Solidity: function MAINTENANCE_MARGIN_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMSession) MAINTENANCEMARGINBPS() (*big.Int, error) {
	return _AppleAMM.Contract.MAINTENANCEMARGINBPS(&_AppleAMM.CallOpts)
}

// MAINTENANCEMARGINBPS is a free data retrieval call binding the contract method 0x5838bff7.
//
// Solidity: function MAINTENANCE_MARGIN_BPS() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) MAINTENANCEMARGINBPS() (*big.Int, error) {
	return _AppleAMM.Contract.MAINTENANCEMARGINBPS(&_AppleAMM.CallOpts)
}

// AppleReserve is a free data retrieval call binding the contract method 0xc70f86dc.
//
// Solidity: function appleReserve() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) AppleReserve(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "appleReserve")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AppleReserve is a free data retrieval call binding the contract method 0xc70f86dc.
//
// Solidity: function appleReserve() view returns(uint256)
func (_AppleAMM *AppleAMMSession) AppleReserve() (*big.Int, error) {
	return _AppleAMM.Contract.AppleReserve(&_AppleAMM.CallOpts)
}

// AppleReserve is a free data retrieval call binding the contract method 0xc70f86dc.
//
// Solidity: function appleReserve() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) AppleReserve() (*big.Int, error) {
	return _AppleAMM.Contract.AppleReserve(&_AppleAMM.CallOpts)
}

// AppleToken is a free data retrieval call binding the contract method 0x12cd255c.
//
// Solidity: function appleToken() view returns(address)
func (_AppleAMM *AppleAMMCaller) AppleToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "appleToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AppleToken is a free data retrieval call binding the contract method 0x12cd255c.
//
// Solidity: function appleToken() view returns(address)
func (_AppleAMM *AppleAMMSession) AppleToken() (common.Address, error) {
	return _AppleAMM.Contract.AppleToken(&_AppleAMM.CallOpts)
}

// AppleToken is a free data retrieval call binding the contract method 0x12cd255c.
//
// Solidity: function appleToken() view returns(address)
func (_AppleAMM *AppleAMMCallerSession) AppleToken() (common.Address, error) {
	return _AppleAMM.Contract.AppleToken(&_AppleAMM.CallOpts)
}

// EthReserve is a free data retrieval call binding the contract method 0xd62ccb3f.
//
// Solidity: function ethReserve() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) EthReserve(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "ethReserve")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EthReserve is a free data retrieval call binding the contract method 0xd62ccb3f.
//
// Solidity: function ethReserve() view returns(uint256)
func (_AppleAMM *AppleAMMSession) EthReserve() (*big.Int, error) {
	return _AppleAMM.Contract.EthReserve(&_AppleAMM.CallOpts)
}

// EthReserve is a free data retrieval call binding the contract method 0xd62ccb3f.
//
// Solidity: function ethReserve() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) EthReserve() (*big.Int, error) {
	return _AppleAMM.Contract.EthReserve(&_AppleAMM.CallOpts)
}

// GetAmountOut is a free data retrieval call binding the contract method 0x054d50d4.
//
// Solidity: function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut) pure returns(uint256 amountOut)
func (_AppleAMM *AppleAMMCaller) GetAmountOut(opts *bind.CallOpts, amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getAmountOut", amountIn, reserveIn, reserveOut)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAmountOut is a free data retrieval call binding the contract method 0x054d50d4.
//
// Solidity: function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut) pure returns(uint256 amountOut)
func (_AppleAMM *AppleAMMSession) GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) (*big.Int, error) {
	return _AppleAMM.Contract.GetAmountOut(&_AppleAMM.CallOpts, amountIn, reserveIn, reserveOut)
}

// GetAmountOut is a free data retrieval call binding the contract method 0x054d50d4.
//
// Solidity: function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut) pure returns(uint256 amountOut)
func (_AppleAMM *AppleAMMCallerSession) GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) (*big.Int, error) {
	return _AppleAMM.Contract.GetAmountOut(&_AppleAMM.CallOpts, amountIn, reserveIn, reserveOut)
}

// GetLPBalance is a free data retrieval call binding the contract method 0xd2258beb.
//
// Solidity: function getLPBalance(address account) view returns(uint256 balance)
func (_AppleAMM *AppleAMMCaller) GetLPBalance(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getLPBalance", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLPBalance is a free data retrieval call binding the contract method 0xd2258beb.
//
// Solidity: function getLPBalance(address account) view returns(uint256 balance)
func (_AppleAMM *AppleAMMSession) GetLPBalance(account common.Address) (*big.Int, error) {
	return _AppleAMM.Contract.GetLPBalance(&_AppleAMM.CallOpts, account)
}

// GetLPBalance is a free data retrieval call binding the contract method 0xd2258beb.
//
// Solidity: function getLPBalance(address account) view returns(uint256 balance)
func (_AppleAMM *AppleAMMCallerSession) GetLPBalance(account common.Address) (*big.Int, error) {
	return _AppleAMM.Contract.GetLPBalance(&_AppleAMM.CallOpts, account)
}

// GetPosition is a free data retrieval call binding the contract method 0x16c19739.
//
// Solidity: function getPosition(address trader) view returns((address,bool,uint256,uint256,uint256,uint256,uint256,uint256) pos)
func (_AppleAMM *AppleAMMCaller) GetPosition(opts *bind.CallOpts, trader common.Address) (AppleAMMPosition, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getPosition", trader)

	if err != nil {
		return *new(AppleAMMPosition), err
	}

	out0 := *abi.ConvertType(out[0], new(AppleAMMPosition)).(*AppleAMMPosition)

	return out0, err

}

// GetPosition is a free data retrieval call binding the contract method 0x16c19739.
//
// Solidity: function getPosition(address trader) view returns((address,bool,uint256,uint256,uint256,uint256,uint256,uint256) pos)
func (_AppleAMM *AppleAMMSession) GetPosition(trader common.Address) (AppleAMMPosition, error) {
	return _AppleAMM.Contract.GetPosition(&_AppleAMM.CallOpts, trader)
}

// GetPosition is a free data retrieval call binding the contract method 0x16c19739.
//
// Solidity: function getPosition(address trader) view returns((address,bool,uint256,uint256,uint256,uint256,uint256,uint256) pos)
func (_AppleAMM *AppleAMMCallerSession) GetPosition(trader common.Address) (AppleAMMPosition, error) {
	return _AppleAMM.Contract.GetPosition(&_AppleAMM.CallOpts, trader)
}

// GetPositionValue is a free data retrieval call binding the contract method 0x1c083f6a.
//
// Solidity: function getPositionValue(address trader) view returns(uint256)
func (_AppleAMM *AppleAMMCaller) GetPositionValue(opts *bind.CallOpts, trader common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getPositionValue", trader)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPositionValue is a free data retrieval call binding the contract method 0x1c083f6a.
//
// Solidity: function getPositionValue(address trader) view returns(uint256)
func (_AppleAMM *AppleAMMSession) GetPositionValue(trader common.Address) (*big.Int, error) {
	return _AppleAMM.Contract.GetPositionValue(&_AppleAMM.CallOpts, trader)
}

// GetPositionValue is a free data retrieval call binding the contract method 0x1c083f6a.
//
// Solidity: function getPositionValue(address trader) view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) GetPositionValue(trader common.Address) (*big.Int, error) {
	return _AppleAMM.Contract.GetPositionValue(&_AppleAMM.CallOpts, trader)
}

// GetPositionValue0 is a free data retrieval call binding the contract method 0x2c51060f.
//
// Solidity: function getPositionValue(address trader, uint256 currentPrice) view returns(uint256)
func (_AppleAMM *AppleAMMCaller) GetPositionValue0(opts *bind.CallOpts, trader common.Address, currentPrice *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getPositionValue0", trader, currentPrice)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPositionValue0 is a free data retrieval call binding the contract method 0x2c51060f.
//
// Solidity: function getPositionValue(address trader, uint256 currentPrice) view returns(uint256)
func (_AppleAMM *AppleAMMSession) GetPositionValue0(trader common.Address, currentPrice *big.Int) (*big.Int, error) {
	return _AppleAMM.Contract.GetPositionValue0(&_AppleAMM.CallOpts, trader, currentPrice)
}

// GetPositionValue0 is a free data retrieval call binding the contract method 0x2c51060f.
//
// Solidity: function getPositionValue(address trader, uint256 currentPrice) view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) GetPositionValue0(trader common.Address, currentPrice *big.Int) (*big.Int, error) {
	return _AppleAMM.Contract.GetPositionValue0(&_AppleAMM.CallOpts, trader, currentPrice)
}

// GetReserves is a free data retrieval call binding the contract method 0x0902f1ac.
//
// Solidity: function getReserves() view returns(uint256 _appleReserve, uint256 _ethReserve)
func (_AppleAMM *AppleAMMCaller) GetReserves(opts *bind.CallOpts) (struct {
	AppleReserve *big.Int
	EthReserve   *big.Int
}, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getReserves")

	outstruct := new(struct {
		AppleReserve *big.Int
		EthReserve   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.AppleReserve = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.EthReserve = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetReserves is a free data retrieval call binding the contract method 0x0902f1ac.
//
// Solidity: function getReserves() view returns(uint256 _appleReserve, uint256 _ethReserve)
func (_AppleAMM *AppleAMMSession) GetReserves() (struct {
	AppleReserve *big.Int
	EthReserve   *big.Int
}, error) {
	return _AppleAMM.Contract.GetReserves(&_AppleAMM.CallOpts)
}

// GetReserves is a free data retrieval call binding the contract method 0x0902f1ac.
//
// Solidity: function getReserves() view returns(uint256 _appleReserve, uint256 _ethReserve)
func (_AppleAMM *AppleAMMCallerSession) GetReserves() (struct {
	AppleReserve *big.Int
	EthReserve   *big.Int
}, error) {
	return _AppleAMM.Contract.GetReserves(&_AppleAMM.CallOpts)
}

// GetSpotPrice is a free data retrieval call binding the contract method 0xdc76fabc.
//
// Solidity: function getSpotPrice() view returns(uint256 price)
func (_AppleAMM *AppleAMMCaller) GetSpotPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getSpotPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSpotPrice is a free data retrieval call binding the contract method 0xdc76fabc.
//
// Solidity: function getSpotPrice() view returns(uint256 price)
func (_AppleAMM *AppleAMMSession) GetSpotPrice() (*big.Int, error) {
	return _AppleAMM.Contract.GetSpotPrice(&_AppleAMM.CallOpts)
}

// GetSpotPrice is a free data retrieval call binding the contract method 0xdc76fabc.
//
// Solidity: function getSpotPrice() view returns(uint256 price)
func (_AppleAMM *AppleAMMCallerSession) GetSpotPrice() (*big.Int, error) {
	return _AppleAMM.Contract.GetSpotPrice(&_AppleAMM.CallOpts)
}

// GetTotalFees is a free data retrieval call binding the contract method 0x626e1ae7.
//
// Solidity: function getTotalFees() view returns(uint256 feesApple, uint256 feesETH)
func (_AppleAMM *AppleAMMCaller) GetTotalFees(opts *bind.CallOpts) (struct {
	FeesApple *big.Int
	FeesETH   *big.Int
}, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "getTotalFees")

	outstruct := new(struct {
		FeesApple *big.Int
		FeesETH   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.FeesApple = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.FeesETH = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetTotalFees is a free data retrieval call binding the contract method 0x626e1ae7.
//
// Solidity: function getTotalFees() view returns(uint256 feesApple, uint256 feesETH)
func (_AppleAMM *AppleAMMSession) GetTotalFees() (struct {
	FeesApple *big.Int
	FeesETH   *big.Int
}, error) {
	return _AppleAMM.Contract.GetTotalFees(&_AppleAMM.CallOpts)
}

// GetTotalFees is a free data retrieval call binding the contract method 0x626e1ae7.
//
// Solidity: function getTotalFees() view returns(uint256 feesApple, uint256 feesETH)
func (_AppleAMM *AppleAMMCallerSession) GetTotalFees() (struct {
	FeesApple *big.Int
	FeesETH   *big.Int
}, error) {
	return _AppleAMM.Contract.GetTotalFees(&_AppleAMM.CallOpts)
}

// IsLiquidatable is a free data retrieval call binding the contract method 0x042e02cf.
//
// Solidity: function isLiquidatable(address trader) view returns(bool liquidatable, uint256 marginRatio)
func (_AppleAMM *AppleAMMCaller) IsLiquidatable(opts *bind.CallOpts, trader common.Address) (struct {
	Liquidatable bool
	MarginRatio  *big.Int
}, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "isLiquidatable", trader)

	outstruct := new(struct {
		Liquidatable bool
		MarginRatio  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Liquidatable = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.MarginRatio = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// IsLiquidatable is a free data retrieval call binding the contract method 0x042e02cf.
//
// Solidity: function isLiquidatable(address trader) view returns(bool liquidatable, uint256 marginRatio)
func (_AppleAMM *AppleAMMSession) IsLiquidatable(trader common.Address) (struct {
	Liquidatable bool
	MarginRatio  *big.Int
}, error) {
	return _AppleAMM.Contract.IsLiquidatable(&_AppleAMM.CallOpts, trader)
}

// IsLiquidatable is a free data retrieval call binding the contract method 0x042e02cf.
//
// Solidity: function isLiquidatable(address trader) view returns(bool liquidatable, uint256 marginRatio)
func (_AppleAMM *AppleAMMCallerSession) IsLiquidatable(trader common.Address) (struct {
	Liquidatable bool
	MarginRatio  *big.Int
}, error) {
	return _AppleAMM.Contract.IsLiquidatable(&_AppleAMM.CallOpts, trader)
}

// LpBalances is a free data retrieval call binding the contract method 0x0b65092d.
//
// Solidity: function lpBalances(address ) view returns(uint256)
func (_AppleAMM *AppleAMMCaller) LpBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "lpBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LpBalances is a free data retrieval call binding the contract method 0x0b65092d.
//
// Solidity: function lpBalances(address ) view returns(uint256)
func (_AppleAMM *AppleAMMSession) LpBalances(arg0 common.Address) (*big.Int, error) {
	return _AppleAMM.Contract.LpBalances(&_AppleAMM.CallOpts, arg0)
}

// LpBalances is a free data retrieval call binding the contract method 0x0b65092d.
//
// Solidity: function lpBalances(address ) view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) LpBalances(arg0 common.Address) (*big.Int, error) {
	return _AppleAMM.Contract.LpBalances(&_AppleAMM.CallOpts, arg0)
}

// Positions is a free data retrieval call binding the contract method 0x55f57510.
//
// Solidity: function positions(address ) view returns(address trader, bool isLong, uint256 collateral, uint256 size, uint256 entryPrice, uint256 leverage, uint256 loanAmount, uint256 timestamp)
func (_AppleAMM *AppleAMMCaller) Positions(opts *bind.CallOpts, arg0 common.Address) (struct {
	Trader     common.Address
	IsLong     bool
	Collateral *big.Int
	Size       *big.Int
	EntryPrice *big.Int
	Leverage   *big.Int
	LoanAmount *big.Int
	Timestamp  *big.Int
}, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "positions", arg0)

	outstruct := new(struct {
		Trader     common.Address
		IsLong     bool
		Collateral *big.Int
		Size       *big.Int
		EntryPrice *big.Int
		Leverage   *big.Int
		LoanAmount *big.Int
		Timestamp  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Trader = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.IsLong = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.Collateral = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Size = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.EntryPrice = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Leverage = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.LoanAmount = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.Timestamp = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Positions is a free data retrieval call binding the contract method 0x55f57510.
//
// Solidity: function positions(address ) view returns(address trader, bool isLong, uint256 collateral, uint256 size, uint256 entryPrice, uint256 leverage, uint256 loanAmount, uint256 timestamp)
func (_AppleAMM *AppleAMMSession) Positions(arg0 common.Address) (struct {
	Trader     common.Address
	IsLong     bool
	Collateral *big.Int
	Size       *big.Int
	EntryPrice *big.Int
	Leverage   *big.Int
	LoanAmount *big.Int
	Timestamp  *big.Int
}, error) {
	return _AppleAMM.Contract.Positions(&_AppleAMM.CallOpts, arg0)
}

// Positions is a free data retrieval call binding the contract method 0x55f57510.
//
// Solidity: function positions(address ) view returns(address trader, bool isLong, uint256 collateral, uint256 size, uint256 entryPrice, uint256 leverage, uint256 loanAmount, uint256 timestamp)
func (_AppleAMM *AppleAMMCallerSession) Positions(arg0 common.Address) (struct {
	Trader     common.Address
	IsLong     bool
	Collateral *big.Int
	Size       *big.Int
	EntryPrice *big.Int
	Leverage   *big.Int
	LoanAmount *big.Int
	Timestamp  *big.Int
}, error) {
	return _AppleAMM.Contract.Positions(&_AppleAMM.CallOpts, arg0)
}

// TotalFeesApple is a free data retrieval call binding the contract method 0x5ca86742.
//
// Solidity: function totalFeesApple() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) TotalFeesApple(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "totalFeesApple")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalFeesApple is a free data retrieval call binding the contract method 0x5ca86742.
//
// Solidity: function totalFeesApple() view returns(uint256)
func (_AppleAMM *AppleAMMSession) TotalFeesApple() (*big.Int, error) {
	return _AppleAMM.Contract.TotalFeesApple(&_AppleAMM.CallOpts)
}

// TotalFeesApple is a free data retrieval call binding the contract method 0x5ca86742.
//
// Solidity: function totalFeesApple() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) TotalFeesApple() (*big.Int, error) {
	return _AppleAMM.Contract.TotalFeesApple(&_AppleAMM.CallOpts)
}

// TotalFeesETH is a free data retrieval call binding the contract method 0x525f85ce.
//
// Solidity: function totalFeesETH() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) TotalFeesETH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "totalFeesETH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalFeesETH is a free data retrieval call binding the contract method 0x525f85ce.
//
// Solidity: function totalFeesETH() view returns(uint256)
func (_AppleAMM *AppleAMMSession) TotalFeesETH() (*big.Int, error) {
	return _AppleAMM.Contract.TotalFeesETH(&_AppleAMM.CallOpts)
}

// TotalFeesETH is a free data retrieval call binding the contract method 0x525f85ce.
//
// Solidity: function totalFeesETH() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) TotalFeesETH() (*big.Int, error) {
	return _AppleAMM.Contract.TotalFeesETH(&_AppleAMM.CallOpts)
}

// TotalLPTokens is a free data retrieval call binding the contract method 0xef6283d9.
//
// Solidity: function totalLPTokens() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) TotalLPTokens(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "totalLPTokens")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalLPTokens is a free data retrieval call binding the contract method 0xef6283d9.
//
// Solidity: function totalLPTokens() view returns(uint256)
func (_AppleAMM *AppleAMMSession) TotalLPTokens() (*big.Int, error) {
	return _AppleAMM.Contract.TotalLPTokens(&_AppleAMM.CallOpts)
}

// TotalLPTokens is a free data retrieval call binding the contract method 0xef6283d9.
//
// Solidity: function totalLPTokens() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) TotalLPTokens() (*big.Int, error) {
	return _AppleAMM.Contract.TotalLPTokens(&_AppleAMM.CallOpts)
}

// TotalLentETH is a free data retrieval call binding the contract method 0x7311d845.
//
// Solidity: function totalLentETH() view returns(uint256)
func (_AppleAMM *AppleAMMCaller) TotalLentETH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppleAMM.contract.Call(opts, &out, "totalLentETH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalLentETH is a free data retrieval call binding the contract method 0x7311d845.
//
// Solidity: function totalLentETH() view returns(uint256)
func (_AppleAMM *AppleAMMSession) TotalLentETH() (*big.Int, error) {
	return _AppleAMM.Contract.TotalLentETH(&_AppleAMM.CallOpts)
}

// TotalLentETH is a free data retrieval call binding the contract method 0x7311d845.
//
// Solidity: function totalLentETH() view returns(uint256)
func (_AppleAMM *AppleAMMCallerSession) TotalLentETH() (*big.Int, error) {
	return _AppleAMM.Contract.TotalLentETH(&_AppleAMM.CallOpts)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x51c6590a.
//
// Solidity: function addLiquidity(uint256 appleAmount) payable returns(uint256 lpTokens)
func (_AppleAMM *AppleAMMTransactor) AddLiquidity(opts *bind.TransactOpts, appleAmount *big.Int) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "addLiquidity", appleAmount)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x51c6590a.
//
// Solidity: function addLiquidity(uint256 appleAmount) payable returns(uint256 lpTokens)
func (_AppleAMM *AppleAMMSession) AddLiquidity(appleAmount *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.AddLiquidity(&_AppleAMM.TransactOpts, appleAmount)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x51c6590a.
//
// Solidity: function addLiquidity(uint256 appleAmount) payable returns(uint256 lpTokens)
func (_AppleAMM *AppleAMMTransactorSession) AddLiquidity(appleAmount *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.AddLiquidity(&_AppleAMM.TransactOpts, appleAmount)
}

// ClosePosition is a paid mutator transaction binding the contract method 0xc393d0e3.
//
// Solidity: function closePosition() returns(uint256 ethOut)
func (_AppleAMM *AppleAMMTransactor) ClosePosition(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "closePosition")
}

// ClosePosition is a paid mutator transaction binding the contract method 0xc393d0e3.
//
// Solidity: function closePosition() returns(uint256 ethOut)
func (_AppleAMM *AppleAMMSession) ClosePosition() (*types.Transaction, error) {
	return _AppleAMM.Contract.ClosePosition(&_AppleAMM.TransactOpts)
}

// ClosePosition is a paid mutator transaction binding the contract method 0xc393d0e3.
//
// Solidity: function closePosition() returns(uint256 ethOut)
func (_AppleAMM *AppleAMMTransactorSession) ClosePosition() (*types.Transaction, error) {
	return _AppleAMM.Contract.ClosePosition(&_AppleAMM.TransactOpts)
}

// Liquidate is a paid mutator transaction binding the contract method 0x2f865568.
//
// Solidity: function liquidate(address trader) returns(bool success)
func (_AppleAMM *AppleAMMTransactor) Liquidate(opts *bind.TransactOpts, trader common.Address) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "liquidate", trader)
}

// Liquidate is a paid mutator transaction binding the contract method 0x2f865568.
//
// Solidity: function liquidate(address trader) returns(bool success)
func (_AppleAMM *AppleAMMSession) Liquidate(trader common.Address) (*types.Transaction, error) {
	return _AppleAMM.Contract.Liquidate(&_AppleAMM.TransactOpts, trader)
}

// Liquidate is a paid mutator transaction binding the contract method 0x2f865568.
//
// Solidity: function liquidate(address trader) returns(bool success)
func (_AppleAMM *AppleAMMTransactorSession) Liquidate(trader common.Address) (*types.Transaction, error) {
	return _AppleAMM.Contract.Liquidate(&_AppleAMM.TransactOpts, trader)
}

// OpenLeveragedPosition is a paid mutator transaction binding the contract method 0xc08f9984.
//
// Solidity: function openLeveragedPosition(uint256 leverage, uint256 minApples) payable returns(uint256 applesOut)
func (_AppleAMM *AppleAMMTransactor) OpenLeveragedPosition(opts *bind.TransactOpts, leverage *big.Int, minApples *big.Int) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "openLeveragedPosition", leverage, minApples)
}

// OpenLeveragedPosition is a paid mutator transaction binding the contract method 0xc08f9984.
//
// Solidity: function openLeveragedPosition(uint256 leverage, uint256 minApples) payable returns(uint256 applesOut)
func (_AppleAMM *AppleAMMSession) OpenLeveragedPosition(leverage *big.Int, minApples *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.OpenLeveragedPosition(&_AppleAMM.TransactOpts, leverage, minApples)
}

// OpenLeveragedPosition is a paid mutator transaction binding the contract method 0xc08f9984.
//
// Solidity: function openLeveragedPosition(uint256 leverage, uint256 minApples) payable returns(uint256 applesOut)
func (_AppleAMM *AppleAMMTransactorSession) OpenLeveragedPosition(leverage *big.Int, minApples *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.OpenLeveragedPosition(&_AppleAMM.TransactOpts, leverage, minApples)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x9c8f9f23.
//
// Solidity: function removeLiquidity(uint256 lpTokenAmount) returns(uint256 appleOut, uint256 ethOut)
func (_AppleAMM *AppleAMMTransactor) RemoveLiquidity(opts *bind.TransactOpts, lpTokenAmount *big.Int) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "removeLiquidity", lpTokenAmount)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x9c8f9f23.
//
// Solidity: function removeLiquidity(uint256 lpTokenAmount) returns(uint256 appleOut, uint256 ethOut)
func (_AppleAMM *AppleAMMSession) RemoveLiquidity(lpTokenAmount *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.RemoveLiquidity(&_AppleAMM.TransactOpts, lpTokenAmount)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x9c8f9f23.
//
// Solidity: function removeLiquidity(uint256 lpTokenAmount) returns(uint256 appleOut, uint256 ethOut)
func (_AppleAMM *AppleAMMTransactorSession) RemoveLiquidity(lpTokenAmount *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.RemoveLiquidity(&_AppleAMM.TransactOpts, lpTokenAmount)
}

// SwapApplesForETH is a paid mutator transaction binding the contract method 0x48af5721.
//
// Solidity: function swapApplesForETH(uint256 appleAmount, uint256 minETH) returns(uint256 ethOut)
func (_AppleAMM *AppleAMMTransactor) SwapApplesForETH(opts *bind.TransactOpts, appleAmount *big.Int, minETH *big.Int) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "swapApplesForETH", appleAmount, minETH)
}

// SwapApplesForETH is a paid mutator transaction binding the contract method 0x48af5721.
//
// Solidity: function swapApplesForETH(uint256 appleAmount, uint256 minETH) returns(uint256 ethOut)
func (_AppleAMM *AppleAMMSession) SwapApplesForETH(appleAmount *big.Int, minETH *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.SwapApplesForETH(&_AppleAMM.TransactOpts, appleAmount, minETH)
}

// SwapApplesForETH is a paid mutator transaction binding the contract method 0x48af5721.
//
// Solidity: function swapApplesForETH(uint256 appleAmount, uint256 minETH) returns(uint256 ethOut)
func (_AppleAMM *AppleAMMTransactorSession) SwapApplesForETH(appleAmount *big.Int, minETH *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.SwapApplesForETH(&_AppleAMM.TransactOpts, appleAmount, minETH)
}

// SwapETHForApples is a paid mutator transaction binding the contract method 0xe26a3c8a.
//
// Solidity: function swapETHForApples(uint256 minApples) payable returns(uint256 applesOut)
func (_AppleAMM *AppleAMMTransactor) SwapETHForApples(opts *bind.TransactOpts, minApples *big.Int) (*types.Transaction, error) {
	return _AppleAMM.contract.Transact(opts, "swapETHForApples", minApples)
}

// SwapETHForApples is a paid mutator transaction binding the contract method 0xe26a3c8a.
//
// Solidity: function swapETHForApples(uint256 minApples) payable returns(uint256 applesOut)
func (_AppleAMM *AppleAMMSession) SwapETHForApples(minApples *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.SwapETHForApples(&_AppleAMM.TransactOpts, minApples)
}

// SwapETHForApples is a paid mutator transaction binding the contract method 0xe26a3c8a.
//
// Solidity: function swapETHForApples(uint256 minApples) payable returns(uint256 applesOut)
func (_AppleAMM *AppleAMMTransactorSession) SwapETHForApples(minApples *big.Int) (*types.Transaction, error) {
	return _AppleAMM.Contract.SwapETHForApples(&_AppleAMM.TransactOpts, minApples)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_AppleAMM *AppleAMMTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppleAMM.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_AppleAMM *AppleAMMSession) Receive() (*types.Transaction, error) {
	return _AppleAMM.Contract.Receive(&_AppleAMM.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_AppleAMM *AppleAMMTransactorSession) Receive() (*types.Transaction, error) {
	return _AppleAMM.Contract.Receive(&_AppleAMM.TransactOpts)
}

// AppleAMMLogLiquidityAddedIterator is returned from FilterLogLiquidityAdded and is used to iterate over the raw logs and unpacked data for LogLiquidityAdded events raised by the AppleAMM contract.
type AppleAMMLogLiquidityAddedIterator struct {
	Event *AppleAMMLogLiquidityAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AppleAMMLogLiquidityAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppleAMMLogLiquidityAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AppleAMMLogLiquidityAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AppleAMMLogLiquidityAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppleAMMLogLiquidityAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppleAMMLogLiquidityAdded represents a LogLiquidityAdded event raised by the AppleAMM contract.
type AppleAMMLogLiquidityAdded struct {
	Provider    common.Address
	AppleAmount *big.Int
	EthAmount   *big.Int
	LpTokens    *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterLogLiquidityAdded is a free log retrieval operation binding the contract event 0x00f00467526ab75114da7a5ec65bb1b6755a6a23d97b65fe8fc1662c920c6044.
//
// Solidity: event LogLiquidityAdded(address indexed provider, uint256 appleAmount, uint256 ethAmount, uint256 lpTokens)
func (_AppleAMM *AppleAMMFilterer) FilterLogLiquidityAdded(opts *bind.FilterOpts, provider []common.Address) (*AppleAMMLogLiquidityAddedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _AppleAMM.contract.FilterLogs(opts, "LogLiquidityAdded", providerRule)
	if err != nil {
		return nil, err
	}
	return &AppleAMMLogLiquidityAddedIterator{contract: _AppleAMM.contract, event: "LogLiquidityAdded", logs: logs, sub: sub}, nil
}

// WatchLogLiquidityAdded is a free log subscription operation binding the contract event 0x00f00467526ab75114da7a5ec65bb1b6755a6a23d97b65fe8fc1662c920c6044.
//
// Solidity: event LogLiquidityAdded(address indexed provider, uint256 appleAmount, uint256 ethAmount, uint256 lpTokens)
func (_AppleAMM *AppleAMMFilterer) WatchLogLiquidityAdded(opts *bind.WatchOpts, sink chan<- *AppleAMMLogLiquidityAdded, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _AppleAMM.contract.WatchLogs(opts, "LogLiquidityAdded", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppleAMMLogLiquidityAdded)
				if err := _AppleAMM.contract.UnpackLog(event, "LogLiquidityAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLogLiquidityAdded is a log parse operation binding the contract event 0x00f00467526ab75114da7a5ec65bb1b6755a6a23d97b65fe8fc1662c920c6044.
//
// Solidity: event LogLiquidityAdded(address indexed provider, uint256 appleAmount, uint256 ethAmount, uint256 lpTokens)
func (_AppleAMM *AppleAMMFilterer) ParseLogLiquidityAdded(log types.Log) (*AppleAMMLogLiquidityAdded, error) {
	event := new(AppleAMMLogLiquidityAdded)
	if err := _AppleAMM.contract.UnpackLog(event, "LogLiquidityAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppleAMMLogLiquidityRemovedIterator is returned from FilterLogLiquidityRemoved and is used to iterate over the raw logs and unpacked data for LogLiquidityRemoved events raised by the AppleAMM contract.
type AppleAMMLogLiquidityRemovedIterator struct {
	Event *AppleAMMLogLiquidityRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AppleAMMLogLiquidityRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppleAMMLogLiquidityRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AppleAMMLogLiquidityRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AppleAMMLogLiquidityRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppleAMMLogLiquidityRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppleAMMLogLiquidityRemoved represents a LogLiquidityRemoved event raised by the AppleAMM contract.
type AppleAMMLogLiquidityRemoved struct {
	Provider    common.Address
	AppleAmount *big.Int
	EthAmount   *big.Int
	LpTokens    *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterLogLiquidityRemoved is a free log retrieval operation binding the contract event 0xf0725bf0beec835ae56762ebed93ca371f2cfef5957bef9e20edd3965aa592ce.
//
// Solidity: event LogLiquidityRemoved(address indexed provider, uint256 appleAmount, uint256 ethAmount, uint256 lpTokens)
func (_AppleAMM *AppleAMMFilterer) FilterLogLiquidityRemoved(opts *bind.FilterOpts, provider []common.Address) (*AppleAMMLogLiquidityRemovedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _AppleAMM.contract.FilterLogs(opts, "LogLiquidityRemoved", providerRule)
	if err != nil {
		return nil, err
	}
	return &AppleAMMLogLiquidityRemovedIterator{contract: _AppleAMM.contract, event: "LogLiquidityRemoved", logs: logs, sub: sub}, nil
}

// WatchLogLiquidityRemoved is a free log subscription operation binding the contract event 0xf0725bf0beec835ae56762ebed93ca371f2cfef5957bef9e20edd3965aa592ce.
//
// Solidity: event LogLiquidityRemoved(address indexed provider, uint256 appleAmount, uint256 ethAmount, uint256 lpTokens)
func (_AppleAMM *AppleAMMFilterer) WatchLogLiquidityRemoved(opts *bind.WatchOpts, sink chan<- *AppleAMMLogLiquidityRemoved, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _AppleAMM.contract.WatchLogs(opts, "LogLiquidityRemoved", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppleAMMLogLiquidityRemoved)
				if err := _AppleAMM.contract.UnpackLog(event, "LogLiquidityRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLogLiquidityRemoved is a log parse operation binding the contract event 0xf0725bf0beec835ae56762ebed93ca371f2cfef5957bef9e20edd3965aa592ce.
//
// Solidity: event LogLiquidityRemoved(address indexed provider, uint256 appleAmount, uint256 ethAmount, uint256 lpTokens)
func (_AppleAMM *AppleAMMFilterer) ParseLogLiquidityRemoved(log types.Log) (*AppleAMMLogLiquidityRemoved, error) {
	event := new(AppleAMMLogLiquidityRemoved)
	if err := _AppleAMM.contract.UnpackLog(event, "LogLiquidityRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppleAMMLogSwapIterator is returned from FilterLogSwap and is used to iterate over the raw logs and unpacked data for LogSwap events raised by the AppleAMM contract.
type AppleAMMLogSwapIterator struct {
	Event *AppleAMMLogSwap // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AppleAMMLogSwapIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppleAMMLogSwap)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AppleAMMLogSwap)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AppleAMMLogSwapIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppleAMMLogSwapIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppleAMMLogSwap represents a LogSwap event raised by the AppleAMM contract.
type AppleAMMLogSwap struct {
	Trader    common.Address
	IsBuy     bool
	AmountIn  *big.Int
	AmountOut *big.Int
	Price     *big.Int
	Fee       *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterLogSwap is a free log retrieval operation binding the contract event 0x11fc2ef3451a45abcd1e04d7185f558c9e73f0093a16a4743c2f429e78d31759.
//
// Solidity: event LogSwap(address indexed trader, bool isBuy, uint256 amountIn, uint256 amountOut, uint256 price, uint256 fee, uint256 timestamp)
func (_AppleAMM *AppleAMMFilterer) FilterLogSwap(opts *bind.FilterOpts, trader []common.Address) (*AppleAMMLogSwapIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _AppleAMM.contract.FilterLogs(opts, "LogSwap", traderRule)
	if err != nil {
		return nil, err
	}
	return &AppleAMMLogSwapIterator{contract: _AppleAMM.contract, event: "LogSwap", logs: logs, sub: sub}, nil
}

// WatchLogSwap is a free log subscription operation binding the contract event 0x11fc2ef3451a45abcd1e04d7185f558c9e73f0093a16a4743c2f429e78d31759.
//
// Solidity: event LogSwap(address indexed trader, bool isBuy, uint256 amountIn, uint256 amountOut, uint256 price, uint256 fee, uint256 timestamp)
func (_AppleAMM *AppleAMMFilterer) WatchLogSwap(opts *bind.WatchOpts, sink chan<- *AppleAMMLogSwap, trader []common.Address) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _AppleAMM.contract.WatchLogs(opts, "LogSwap", traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppleAMMLogSwap)
				if err := _AppleAMM.contract.UnpackLog(event, "LogSwap", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLogSwap is a log parse operation binding the contract event 0x11fc2ef3451a45abcd1e04d7185f558c9e73f0093a16a4743c2f429e78d31759.
//
// Solidity: event LogSwap(address indexed trader, bool isBuy, uint256 amountIn, uint256 amountOut, uint256 price, uint256 fee, uint256 timestamp)
func (_AppleAMM *AppleAMMFilterer) ParseLogSwap(log types.Log) (*AppleAMMLogSwap, error) {
	event := new(AppleAMMLogSwap)
	if err := _AppleAMM.contract.UnpackLog(event, "LogSwap", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppleAMMPositionClosedIterator is returned from FilterPositionClosed and is used to iterate over the raw logs and unpacked data for PositionClosed events raised by the AppleAMM contract.
type AppleAMMPositionClosedIterator struct {
	Event *AppleAMMPositionClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AppleAMMPositionClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppleAMMPositionClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AppleAMMPositionClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AppleAMMPositionClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppleAMMPositionClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppleAMMPositionClosed represents a PositionClosed event raised by the AppleAMM contract.
type AppleAMMPositionClosed struct {
	Trader             common.Address
	CollateralReturned *big.Int
	Pnl                *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterPositionClosed is a free log retrieval operation binding the contract event 0xc8469f8998427d0814b7efed804617bd8c43dfdfb3c6780d28a03ecf361fb698.
//
// Solidity: event PositionClosed(address indexed trader, uint256 collateralReturned, uint256 pnl)
func (_AppleAMM *AppleAMMFilterer) FilterPositionClosed(opts *bind.FilterOpts, trader []common.Address) (*AppleAMMPositionClosedIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _AppleAMM.contract.FilterLogs(opts, "PositionClosed", traderRule)
	if err != nil {
		return nil, err
	}
	return &AppleAMMPositionClosedIterator{contract: _AppleAMM.contract, event: "PositionClosed", logs: logs, sub: sub}, nil
}

// WatchPositionClosed is a free log subscription operation binding the contract event 0xc8469f8998427d0814b7efed804617bd8c43dfdfb3c6780d28a03ecf361fb698.
//
// Solidity: event PositionClosed(address indexed trader, uint256 collateralReturned, uint256 pnl)
func (_AppleAMM *AppleAMMFilterer) WatchPositionClosed(opts *bind.WatchOpts, sink chan<- *AppleAMMPositionClosed, trader []common.Address) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _AppleAMM.contract.WatchLogs(opts, "PositionClosed", traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppleAMMPositionClosed)
				if err := _AppleAMM.contract.UnpackLog(event, "PositionClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePositionClosed is a log parse operation binding the contract event 0xc8469f8998427d0814b7efed804617bd8c43dfdfb3c6780d28a03ecf361fb698.
//
// Solidity: event PositionClosed(address indexed trader, uint256 collateralReturned, uint256 pnl)
func (_AppleAMM *AppleAMMFilterer) ParsePositionClosed(log types.Log) (*AppleAMMPositionClosed, error) {
	event := new(AppleAMMPositionClosed)
	if err := _AppleAMM.contract.UnpackLog(event, "PositionClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppleAMMPositionLiquidatedIterator is returned from FilterPositionLiquidated and is used to iterate over the raw logs and unpacked data for PositionLiquidated events raised by the AppleAMM contract.
type AppleAMMPositionLiquidatedIterator struct {
	Event *AppleAMMPositionLiquidated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AppleAMMPositionLiquidatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppleAMMPositionLiquidated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AppleAMMPositionLiquidated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AppleAMMPositionLiquidatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppleAMMPositionLiquidatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppleAMMPositionLiquidated represents a PositionLiquidated event raised by the AppleAMM contract.
type AppleAMMPositionLiquidated struct {
	Trader              common.Address
	Liquidator          common.Address
	Collateral          *big.Int
	LiquidationFee      *big.Int
	RemainingCollateral *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterPositionLiquidated is a free log retrieval operation binding the contract event 0x3ecbb371880c45545c63ac02e2364778ac379d2c14519773964068ef51bb454a.
//
// Solidity: event PositionLiquidated(address indexed trader, address indexed liquidator, uint256 collateral, uint256 liquidationFee, uint256 remainingCollateral)
func (_AppleAMM *AppleAMMFilterer) FilterPositionLiquidated(opts *bind.FilterOpts, trader []common.Address, liquidator []common.Address) (*AppleAMMPositionLiquidatedIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}
	var liquidatorRule []interface{}
	for _, liquidatorItem := range liquidator {
		liquidatorRule = append(liquidatorRule, liquidatorItem)
	}

	logs, sub, err := _AppleAMM.contract.FilterLogs(opts, "PositionLiquidated", traderRule, liquidatorRule)
	if err != nil {
		return nil, err
	}
	return &AppleAMMPositionLiquidatedIterator{contract: _AppleAMM.contract, event: "PositionLiquidated", logs: logs, sub: sub}, nil
}

// WatchPositionLiquidated is a free log subscription operation binding the contract event 0x3ecbb371880c45545c63ac02e2364778ac379d2c14519773964068ef51bb454a.
//
// Solidity: event PositionLiquidated(address indexed trader, address indexed liquidator, uint256 collateral, uint256 liquidationFee, uint256 remainingCollateral)
func (_AppleAMM *AppleAMMFilterer) WatchPositionLiquidated(opts *bind.WatchOpts, sink chan<- *AppleAMMPositionLiquidated, trader []common.Address, liquidator []common.Address) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}
	var liquidatorRule []interface{}
	for _, liquidatorItem := range liquidator {
		liquidatorRule = append(liquidatorRule, liquidatorItem)
	}

	logs, sub, err := _AppleAMM.contract.WatchLogs(opts, "PositionLiquidated", traderRule, liquidatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppleAMMPositionLiquidated)
				if err := _AppleAMM.contract.UnpackLog(event, "PositionLiquidated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePositionLiquidated is a log parse operation binding the contract event 0x3ecbb371880c45545c63ac02e2364778ac379d2c14519773964068ef51bb454a.
//
// Solidity: event PositionLiquidated(address indexed trader, address indexed liquidator, uint256 collateral, uint256 liquidationFee, uint256 remainingCollateral)
func (_AppleAMM *AppleAMMFilterer) ParsePositionLiquidated(log types.Log) (*AppleAMMPositionLiquidated, error) {
	event := new(AppleAMMPositionLiquidated)
	if err := _AppleAMM.contract.UnpackLog(event, "PositionLiquidated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppleAMMPositionOpenedIterator is returned from FilterPositionOpened and is used to iterate over the raw logs and unpacked data for PositionOpened events raised by the AppleAMM contract.
type AppleAMMPositionOpenedIterator struct {
	Event *AppleAMMPositionOpened // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AppleAMMPositionOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppleAMMPositionOpened)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AppleAMMPositionOpened)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AppleAMMPositionOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppleAMMPositionOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppleAMMPositionOpened represents a PositionOpened event raised by the AppleAMM contract.
type AppleAMMPositionOpened struct {
	Trader     common.Address
	IsLong     bool
	Collateral *big.Int
	Size       *big.Int
	EntryPrice *big.Int
	Leverage   *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPositionOpened is a free log retrieval operation binding the contract event 0x0e8f4b81343afa64bc4ddcfbf2d80e883e0ddfc6663c400ca09da13d1d28d588.
//
// Solidity: event PositionOpened(address indexed trader, bool isLong, uint256 collateral, uint256 size, uint256 entryPrice, uint256 leverage)
func (_AppleAMM *AppleAMMFilterer) FilterPositionOpened(opts *bind.FilterOpts, trader []common.Address) (*AppleAMMPositionOpenedIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _AppleAMM.contract.FilterLogs(opts, "PositionOpened", traderRule)
	if err != nil {
		return nil, err
	}
	return &AppleAMMPositionOpenedIterator{contract: _AppleAMM.contract, event: "PositionOpened", logs: logs, sub: sub}, nil
}

// WatchPositionOpened is a free log subscription operation binding the contract event 0x0e8f4b81343afa64bc4ddcfbf2d80e883e0ddfc6663c400ca09da13d1d28d588.
//
// Solidity: event PositionOpened(address indexed trader, bool isLong, uint256 collateral, uint256 size, uint256 entryPrice, uint256 leverage)
func (_AppleAMM *AppleAMMFilterer) WatchPositionOpened(opts *bind.WatchOpts, sink chan<- *AppleAMMPositionOpened, trader []common.Address) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _AppleAMM.contract.WatchLogs(opts, "PositionOpened", traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppleAMMPositionOpened)
				if err := _AppleAMM.contract.UnpackLog(event, "PositionOpened", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePositionOpened is a log parse operation binding the contract event 0x0e8f4b81343afa64bc4ddcfbf2d80e883e0ddfc6663c400ca09da13d1d28d588.
//
// Solidity: event PositionOpened(address indexed trader, bool isLong, uint256 collateral, uint256 size, uint256 entryPrice, uint256 leverage)
func (_AppleAMM *AppleAMMFilterer) ParsePositionOpened(log types.Log) (*AppleAMMPositionOpened, error) {
	event := new(AppleAMMPositionOpened)
	if err := _AppleAMM.contract.UnpackLog(event, "PositionOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
