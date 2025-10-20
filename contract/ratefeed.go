// Code generated via abigen V2 - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = bytes.Equal
	_ = errors.New
	_ = big.NewInt
	_ = common.Big1
	_ = types.BloomLookup
	_ = abi.ConvertType
)

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"sub\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"div\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"mul\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"add\",\"type\":\"uint256\"}],\"name\":\"RateUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"getRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"sub\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"div\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"mul\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"add\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sub\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"div\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"mul\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"add\",\"type\":\"uint256\"}],\"name\":\"setRate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	ID:  "Contract",
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	abi abi.ABI
}

// NewContract creates a new instance of Contract.
func NewContract() *Contract {
	parsed, err := ContractMetaData.ParseABI()
	if err != nil {
		panic(errors.New("invalid ABI: " + err.Error()))
	}
	return &Contract{abi: *parsed}
}

// Instance creates a wrapper for a deployed contract instance at the given address.
// Use this to create the instance object passed to abigen v2 library functions Call, Transact, etc.
func (c *Contract) Instance(backend bind.ContractBackend, addr common.Address) *bind.BoundContract {
	return bind.NewBoundContract(addr, c.abi, backend, backend, backend)
}

// PackGetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x679aefce.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function getRate() view returns(uint256 sub, uint256 div, uint256 mul, uint256 add)
func (contract *Contract) PackGetRate() []byte {
	enc, err := contract.abi.Pack("getRate")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackGetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x679aefce.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function getRate() view returns(uint256 sub, uint256 div, uint256 mul, uint256 add)
func (contract *Contract) TryPackGetRate() ([]byte, error) {
	return contract.abi.Pack("getRate")
}

// GetRateOutput serves as a container for the return parameters of contract
// method GetRate.
type GetRateOutput struct {
	Sub *big.Int
	Div *big.Int
	Mul *big.Int
	Add *big.Int
}

// UnpackGetRate is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x679aefce.
//
// Solidity: function getRate() view returns(uint256 sub, uint256 div, uint256 mul, uint256 add)
func (contract *Contract) UnpackGetRate(data []byte) (GetRateOutput, error) {
	out, err := contract.abi.Unpack("getRate", data)
	outstruct := new(GetRateOutput)
	if err != nil {
		return *outstruct, err
	}
	outstruct.Sub = abi.ConvertType(out[0], new(big.Int)).(*big.Int)
	outstruct.Div = abi.ConvertType(out[1], new(big.Int)).(*big.Int)
	outstruct.Mul = abi.ConvertType(out[2], new(big.Int)).(*big.Int)
	outstruct.Add = abi.ConvertType(out[3], new(big.Int)).(*big.Int)
	return *outstruct, nil
}

// PackOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x8da5cb5b.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function owner() view returns(address)
func (contract *Contract) PackOwner() []byte {
	enc, err := contract.abi.Pack("owner")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x8da5cb5b.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function owner() view returns(address)
func (contract *Contract) TryPackOwner() ([]byte, error) {
	return contract.abi.Pack("owner")
}

// UnpackOwner is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (contract *Contract) UnpackOwner(data []byte) (common.Address, error) {
	out, err := contract.abi.Unpack("owner", data)
	if err != nil {
		return *new(common.Address), err
	}
	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	return out0, nil
}

// PackSetOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x13af4035.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function setOwner(address addr) returns()
func (contract *Contract) PackSetOwner(addr common.Address) []byte {
	enc, err := contract.abi.Pack("setOwner", addr)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackSetOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x13af4035.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function setOwner(address addr) returns()
func (contract *Contract) TryPackSetOwner(addr common.Address) ([]byte, error) {
	return contract.abi.Pack("setOwner", addr)
}

// PackSetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xa3d7a9e4.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function setRate(uint256 sub, uint256 div, uint256 mul, uint256 add) returns()
func (contract *Contract) PackSetRate(sub *big.Int, div *big.Int, mul *big.Int, add *big.Int) []byte {
	enc, err := contract.abi.Pack("setRate", sub, div, mul, add)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackSetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xa3d7a9e4.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function setRate(uint256 sub, uint256 div, uint256 mul, uint256 add) returns()
func (contract *Contract) TryPackSetRate(sub *big.Int, div *big.Int, mul *big.Int, add *big.Int) ([]byte, error) {
	return contract.abi.Pack("setRate", sub, div, mul, add)
}

// ContractOwnerChanged represents a OwnerChanged event raised by the Contract contract.
type ContractOwnerChanged struct {
	OldOwner common.Address
	NewOwner common.Address
	Raw      *types.Log // Blockchain specific contextual infos
}

const ContractOwnerChangedEventName = "OwnerChanged"

// ContractEventName returns the user-defined event name.
func (ContractOwnerChanged) ContractEventName() string {
	return ContractOwnerChangedEventName
}

// UnpackOwnerChangedEvent is the Go binding that unpacks the event data emitted
// by contract.
//
// Solidity: event OwnerChanged(address indexed oldOwner, address indexed newOwner)
func (contract *Contract) UnpackOwnerChangedEvent(log *types.Log) (*ContractOwnerChanged, error) {
	event := "OwnerChanged"
	if len(log.Topics) == 0 || log.Topics[0] != contract.abi.Events[event].ID {
		return nil, errors.New("event signature mismatch")
	}
	out := new(ContractOwnerChanged)
	if len(log.Data) > 0 {
		if err := contract.abi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return nil, err
		}
	}
	var indexed abi.Arguments
	for _, arg := range contract.abi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	if err := abi.ParseTopics(out, indexed, log.Topics[1:]); err != nil {
		return nil, err
	}
	out.Raw = log
	return out, nil
}

// ContractRateUpdated represents a RateUpdated event raised by the Contract contract.
type ContractRateUpdated struct {
	Sub *big.Int
	Div *big.Int
	Mul *big.Int
	Add *big.Int
	Raw *types.Log // Blockchain specific contextual infos
}

const ContractRateUpdatedEventName = "RateUpdated"

// ContractEventName returns the user-defined event name.
func (ContractRateUpdated) ContractEventName() string {
	return ContractRateUpdatedEventName
}

// UnpackRateUpdatedEvent is the Go binding that unpacks the event data emitted
// by contract.
//
// Solidity: event RateUpdated(uint256 sub, uint256 div, uint256 mul, uint256 add)
func (contract *Contract) UnpackRateUpdatedEvent(log *types.Log) (*ContractRateUpdated, error) {
	event := "RateUpdated"
	if len(log.Topics) == 0 || log.Topics[0] != contract.abi.Events[event].ID {
		return nil, errors.New("event signature mismatch")
	}
	out := new(ContractRateUpdated)
	if len(log.Data) > 0 {
		if err := contract.abi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return nil, err
		}
	}
	var indexed abi.Arguments
	for _, arg := range contract.abi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	if err := abi.ParseTopics(out, indexed, log.Topics[1:]); err != nil {
		return nil, err
	}
	out.Raw = log
	return out, nil
}
