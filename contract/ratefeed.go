// Code generated via abigen V2 - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ratefeed

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

// RatefeedMetaData contains all meta data concerning the Ratefeed contract.
var RatefeedMetaData = bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"sub\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"div\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"mul\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"add\",\"type\":\"uint256\"}],\"name\":\"RateUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"add\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"div\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"_sub\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_div\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_mul\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_add\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mul\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"setOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_sub\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_div\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_mul\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_add\",\"type\":\"uint256\"}],\"name\":\"setRate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sub\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	ID:  "Ratefeed",
}

// Ratefeed is an auto generated Go binding around an Ethereum contract.
type Ratefeed struct {
	abi abi.ABI
}

// NewRatefeed creates a new instance of Ratefeed.
func NewRatefeed() *Ratefeed {
	parsed, err := RatefeedMetaData.ParseABI()
	if err != nil {
		panic(errors.New("invalid ABI: " + err.Error()))
	}
	return &Ratefeed{abi: *parsed}
}

// Instance creates a wrapper for a deployed contract instance at the given address.
// Use this to create the instance object passed to abigen v2 library functions Call, Transact, etc.
func (c *Ratefeed) Instance(backend bind.ContractBackend, addr common.Address) *bind.BoundContract {
	return bind.NewBoundContract(addr, c.abi, backend, backend, backend)
}

// PackAdd is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x4f2be91f.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function add() view returns(uint256)
func (ratefeed *Ratefeed) PackAdd() []byte {
	enc, err := ratefeed.abi.Pack("add")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackAdd is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x4f2be91f.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function add() view returns(uint256)
func (ratefeed *Ratefeed) TryPackAdd() ([]byte, error) {
	return ratefeed.abi.Pack("add")
}

// UnpackAdd is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x4f2be91f.
//
// Solidity: function add() view returns(uint256)
func (ratefeed *Ratefeed) UnpackAdd(data []byte) (*big.Int, error) {
	out, err := ratefeed.abi.Unpack("add", data)
	if err != nil {
		return new(big.Int), err
	}
	out0 := abi.ConvertType(out[0], new(big.Int)).(*big.Int)
	return out0, nil
}

// PackDiv is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xf9fa48c3.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function div() view returns(uint256)
func (ratefeed *Ratefeed) PackDiv() []byte {
	enc, err := ratefeed.abi.Pack("div")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackDiv is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xf9fa48c3.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function div() view returns(uint256)
func (ratefeed *Ratefeed) TryPackDiv() ([]byte, error) {
	return ratefeed.abi.Pack("div")
}

// UnpackDiv is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0xf9fa48c3.
//
// Solidity: function div() view returns(uint256)
func (ratefeed *Ratefeed) UnpackDiv(data []byte) (*big.Int, error) {
	out, err := ratefeed.abi.Unpack("div", data)
	if err != nil {
		return new(big.Int), err
	}
	out0 := abi.ConvertType(out[0], new(big.Int)).(*big.Int)
	return out0, nil
}

// PackGetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x679aefce.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function getRate() view returns(uint256 _sub, uint256 _div, uint256 _mul, uint256 _add)
func (ratefeed *Ratefeed) PackGetRate() []byte {
	enc, err := ratefeed.abi.Pack("getRate")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackGetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x679aefce.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function getRate() view returns(uint256 _sub, uint256 _div, uint256 _mul, uint256 _add)
func (ratefeed *Ratefeed) TryPackGetRate() ([]byte, error) {
	return ratefeed.abi.Pack("getRate")
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
// Solidity: function getRate() view returns(uint256 _sub, uint256 _div, uint256 _mul, uint256 _add)
func (ratefeed *Ratefeed) UnpackGetRate(data []byte) (GetRateOutput, error) {
	out, err := ratefeed.abi.Unpack("getRate", data)
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

// PackMul is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x58931c46.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function mul() view returns(uint256)
func (ratefeed *Ratefeed) PackMul() []byte {
	enc, err := ratefeed.abi.Pack("mul")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackMul is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x58931c46.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function mul() view returns(uint256)
func (ratefeed *Ratefeed) TryPackMul() ([]byte, error) {
	return ratefeed.abi.Pack("mul")
}

// UnpackMul is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x58931c46.
//
// Solidity: function mul() view returns(uint256)
func (ratefeed *Ratefeed) UnpackMul(data []byte) (*big.Int, error) {
	out, err := ratefeed.abi.Unpack("mul", data)
	if err != nil {
		return new(big.Int), err
	}
	out0 := abi.ConvertType(out[0], new(big.Int)).(*big.Int)
	return out0, nil
}

// PackOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x8da5cb5b.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function owner() view returns(address)
func (ratefeed *Ratefeed) PackOwner() []byte {
	enc, err := ratefeed.abi.Pack("owner")
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
func (ratefeed *Ratefeed) TryPackOwner() ([]byte, error) {
	return ratefeed.abi.Pack("owner")
}

// UnpackOwner is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (ratefeed *Ratefeed) UnpackOwner(data []byte) (common.Address, error) {
	out, err := ratefeed.abi.Unpack("owner", data)
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
// Solidity: function setOwner(address _addr) returns()
func (ratefeed *Ratefeed) PackSetOwner(addr common.Address) []byte {
	enc, err := ratefeed.abi.Pack("setOwner", addr)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackSetOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x13af4035.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function setOwner(address _addr) returns()
func (ratefeed *Ratefeed) TryPackSetOwner(addr common.Address) ([]byte, error) {
	return ratefeed.abi.Pack("setOwner", addr)
}

// PackSetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xa3d7a9e4.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function setRate(uint256 _sub, uint256 _div, uint256 _mul, uint256 _add) returns()
func (ratefeed *Ratefeed) PackSetRate(sub *big.Int, div *big.Int, mul *big.Int, add *big.Int) []byte {
	enc, err := ratefeed.abi.Pack("setRate", sub, div, mul, add)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackSetRate is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xa3d7a9e4.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function setRate(uint256 _sub, uint256 _div, uint256 _mul, uint256 _add) returns()
func (ratefeed *Ratefeed) TryPackSetRate(sub *big.Int, div *big.Int, mul *big.Int, add *big.Int) ([]byte, error) {
	return ratefeed.abi.Pack("setRate", sub, div, mul, add)
}

// PackSub is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xc54124be.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function sub() view returns(uint256)
func (ratefeed *Ratefeed) PackSub() []byte {
	enc, err := ratefeed.abi.Pack("sub")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackSub is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xc54124be.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function sub() view returns(uint256)
func (ratefeed *Ratefeed) TryPackSub() ([]byte, error) {
	return ratefeed.abi.Pack("sub")
}

// UnpackSub is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0xc54124be.
//
// Solidity: function sub() view returns(uint256)
func (ratefeed *Ratefeed) UnpackSub(data []byte) (*big.Int, error) {
	out, err := ratefeed.abi.Unpack("sub", data)
	if err != nil {
		return new(big.Int), err
	}
	out0 := abi.ConvertType(out[0], new(big.Int)).(*big.Int)
	return out0, nil
}

// RatefeedOwnerChanged represents a OwnerChanged event raised by the Ratefeed contract.
type RatefeedOwnerChanged struct {
	OldOwner common.Address
	NewOwner common.Address
	Raw      *types.Log // Blockchain specific contextual infos
}

const RatefeedOwnerChangedEventName = "OwnerChanged"

// ContractEventName returns the user-defined event name.
func (RatefeedOwnerChanged) ContractEventName() string {
	return RatefeedOwnerChangedEventName
}

// UnpackOwnerChangedEvent is the Go binding that unpacks the event data emitted
// by contract.
//
// Solidity: event OwnerChanged(address indexed oldOwner, address indexed newOwner)
func (ratefeed *Ratefeed) UnpackOwnerChangedEvent(log *types.Log) (*RatefeedOwnerChanged, error) {
	event := "OwnerChanged"
	if len(log.Topics) == 0 || log.Topics[0] != ratefeed.abi.Events[event].ID {
		return nil, errors.New("event signature mismatch")
	}
	out := new(RatefeedOwnerChanged)
	if len(log.Data) > 0 {
		if err := ratefeed.abi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return nil, err
		}
	}
	var indexed abi.Arguments
	for _, arg := range ratefeed.abi.Events[event].Inputs {
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

// RatefeedRateUpdated represents a RateUpdated event raised by the Ratefeed contract.
type RatefeedRateUpdated struct {
	Sub *big.Int
	Div *big.Int
	Mul *big.Int
	Add *big.Int
	Raw *types.Log // Blockchain specific contextual infos
}

const RatefeedRateUpdatedEventName = "RateUpdated"

// ContractEventName returns the user-defined event name.
func (RatefeedRateUpdated) ContractEventName() string {
	return RatefeedRateUpdatedEventName
}

// UnpackRateUpdatedEvent is the Go binding that unpacks the event data emitted
// by contract.
//
// Solidity: event RateUpdated(uint256 sub, uint256 div, uint256 mul, uint256 add)
func (ratefeed *Ratefeed) UnpackRateUpdatedEvent(log *types.Log) (*RatefeedRateUpdated, error) {
	event := "RateUpdated"
	if len(log.Topics) == 0 || log.Topics[0] != ratefeed.abi.Events[event].ID {
		return nil, errors.New("event signature mismatch")
	}
	out := new(RatefeedRateUpdated)
	if len(log.Data) > 0 {
		if err := ratefeed.abi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return nil, err
		}
	}
	var indexed abi.Arguments
	for _, arg := range ratefeed.abi.Events[event].Inputs {
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
