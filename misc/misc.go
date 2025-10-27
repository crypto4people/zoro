package misc

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

func WeiToEth(wei *big.Int) string {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18)).Text('f', 18)
}

func WeiToGwei(wei *big.Int) string {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e9)).Text('f', 2)
}

func SliceToSet[T comparable](slice []T) map[T]struct{} {
	mapp := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		mapp[v] = struct{}{}
	}
	return mapp
}

func ParseEthAddress(str string) (addr common.Address, err error) {
	err = addr.UnmarshalText([]byte(str))
	return addr, err
}

func ParseEthHash(str string) (hash common.Hash, err error) {
	err = hash.UnmarshalText([]byte(str))
	return hash, err
}

func ParseEthTransaction(b []byte) (*types.Transaction, error) {
	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(b)
	return tx, err
}

func ParseBigInt(str string) (*big.Int, error) {
	bi, ok := new(big.Int).SetString(str, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse big int %q", str)
	}
	return bi, nil
}

func ParseUint256(str string) (*uint256.Int, error) {
	u256, err := uint256.FromDecimal(str)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uint256 %q", str)
	}
	return u256, nil
}

func FilterNonReceive(tx *types.Transaction, chain *big.Int, receivers map[common.Address]struct{}) bool {
	to := tx.To()
	return to == nil ||
		tx.ChainId().Cmp(chain) != 0 ||
		tx.Value().Sign() <= 0 ||
		len(tx.Data()) > 0 ||
		(len(receivers) > 0 && !setContains(receivers, *to))
}

func setContains[T comparable](mapp map[T]struct{}, v T) bool {
	_, exists := mapp[v]
	return exists
}
