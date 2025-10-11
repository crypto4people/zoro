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
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnerAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"sub\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"div\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"mul\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"add\",\"type\":\"uint256\"}],\"name\":\"ValuesUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"addOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValues\",\"outputs\":[{\"internalType\":\"uint256[4]\",\"name\":\"\",\"type\":\"uint256[4]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sub\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"div\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"mul\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"add\",\"type\":\"uint256\"}],\"name\":\"setFeed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// PackAddOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x7065cb48.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function addOwner(address _newOwner) returns()
func (ratefeed *Ratefeed) PackAddOwner(newOwner common.Address) []byte {
	enc, err := ratefeed.abi.Pack("addOwner", newOwner)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackAddOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x7065cb48.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function addOwner(address _newOwner) returns()
func (ratefeed *Ratefeed) TryPackAddOwner(newOwner common.Address) ([]byte, error) {
	return ratefeed.abi.Pack("addOwner", newOwner)
}

// PackGetValues is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x19eb4a90.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function getValues() view returns(uint256[4])
func (ratefeed *Ratefeed) PackGetValues() []byte {
	enc, err := ratefeed.abi.Pack("getValues")
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackGetValues is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x19eb4a90.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function getValues() view returns(uint256[4])
func (ratefeed *Ratefeed) TryPackGetValues() ([]byte, error) {
	return ratefeed.abi.Pack("getValues")
}

// UnpackGetValues is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x19eb4a90.
//
// Solidity: function getValues() view returns(uint256[4])
func (ratefeed *Ratefeed) UnpackGetValues(data []byte) ([4]*big.Int, error) {
	out, err := ratefeed.abi.Unpack("getValues", data)
	if err != nil {
		return *new([4]*big.Int), err
	}
	out0 := *abi.ConvertType(out[0], new([4]*big.Int)).(*[4]*big.Int)
	return out0, nil
}

// PackIsOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x2f54bf6e.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function isOwner(address _addr) view returns(bool)
func (ratefeed *Ratefeed) PackIsOwner(addr common.Address) []byte {
	enc, err := ratefeed.abi.Pack("isOwner", addr)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackIsOwner is the Go binding used to pack the parameters required for calling
// the contract method with ID 0x2f54bf6e.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function isOwner(address _addr) view returns(bool)
func (ratefeed *Ratefeed) TryPackIsOwner(addr common.Address) ([]byte, error) {
	return ratefeed.abi.Pack("isOwner", addr)
}

// UnpackIsOwner is the Go binding that unpacks the parameters returned
// from invoking the contract method with ID 0x2f54bf6e.
//
// Solidity: function isOwner(address _addr) view returns(bool)
func (ratefeed *Ratefeed) UnpackIsOwner(data []byte) (bool, error) {
	out, err := ratefeed.abi.Unpack("isOwner", data)
	if err != nil {
		return *new(bool), err
	}
	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	return out0, nil
}

// PackSetFeed is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xd3302ffb.  This method will panic if any
// invalid/nil inputs are passed.
//
// Solidity: function setFeed(uint256 sub, uint256 div, uint256 mul, uint256 add) returns()
func (ratefeed *Ratefeed) PackSetFeed(sub *big.Int, div *big.Int, mul *big.Int, add *big.Int) []byte {
	enc, err := ratefeed.abi.Pack("setFeed", sub, div, mul, add)
	if err != nil {
		panic(err)
	}
	return enc
}

// TryPackSetFeed is the Go binding used to pack the parameters required for calling
// the contract method with ID 0xd3302ffb.  This method will return an error
// if any inputs are invalid/nil.
//
// Solidity: function setFeed(uint256 sub, uint256 div, uint256 mul, uint256 add) returns()
func (ratefeed *Ratefeed) TryPackSetFeed(sub *big.Int, div *big.Int, mul *big.Int, add *big.Int) ([]byte, error) {
	return ratefeed.abi.Pack("setFeed", sub, div, mul, add)
}

// RatefeedOwnerAdded represents a OwnerAdded event raised by the Ratefeed contract.
type RatefeedOwnerAdded struct {
	NewOwner common.Address
	Raw      *types.Log // Blockchain specific contextual infos
}

const RatefeedOwnerAddedEventName = "OwnerAdded"

// ContractEventName returns the user-defined event name.
func (RatefeedOwnerAdded) ContractEventName() string {
	return RatefeedOwnerAddedEventName
}

// UnpackOwnerAddedEvent is the Go binding that unpacks the event data emitted
// by contract.
//
// Solidity: event OwnerAdded(address indexed newOwner)
func (ratefeed *Ratefeed) UnpackOwnerAddedEvent(log *types.Log) (*RatefeedOwnerAdded, error) {
	event := "OwnerAdded"
	if len(log.Topics) == 0 || log.Topics[0] != ratefeed.abi.Events[event].ID {
		return nil, errors.New("event signature mismatch")
	}
	out := new(RatefeedOwnerAdded)
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

// RatefeedValuesUpdated represents a ValuesUpdated event raised by the Ratefeed contract.
type RatefeedValuesUpdated struct {
	Sub *big.Int
	Div *big.Int
	Mul *big.Int
	Add *big.Int
	Raw *types.Log // Blockchain specific contextual infos
}

const RatefeedValuesUpdatedEventName = "ValuesUpdated"

// ContractEventName returns the user-defined event name.
func (RatefeedValuesUpdated) ContractEventName() string {
	return RatefeedValuesUpdatedEventName
}

// UnpackValuesUpdatedEvent is the Go binding that unpacks the event data emitted
// by contract.
//
// Solidity: event ValuesUpdated(uint256 sub, uint256 div, uint256 mul, uint256 add)
func (ratefeed *Ratefeed) UnpackValuesUpdatedEvent(log *types.Log) (*RatefeedValuesUpdated, error) {
	event := "ValuesUpdated"
	if len(log.Topics) == 0 || log.Topics[0] != ratefeed.abi.Events[event].ID {
		return nil, errors.New("event signature mismatch")
	}
	out := new(RatefeedValuesUpdated)
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
