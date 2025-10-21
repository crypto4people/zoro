package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var contractABI = NewContract()
var rateChangedId = contractABI.abi.Events["RateUpdated"].ID

func ABI() *Contract {
	return contractABI
}

func RateChangedID() common.Hash {
	return rateChangedId
}

func CallGetRate(ctx context.Context, client *ethclient.Client, block int64, contract common.Address) (*GetRateOutput, error) {
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &contract,
		Data: contractABI.PackGetRate(),
	}, big.NewInt(block))
	if err != nil {
		return nil, err
	}
	out, err := contractABI.UnpackGetRate(result)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
