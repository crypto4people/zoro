package misc

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func SetContains[T comparable](mapp map[T]struct{}, v T) bool {
	_, exists := mapp[v]
	return exists
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

func FilterNonReceive(tx *types.Transaction, chain string, receivers map[common.Address]struct{}) bool {
	to := tx.To()
	return to == nil || tx.ChainId().Text(10) != chain || tx.Value().Sign() <= 0 || len(tx.Data()) > 0 || !SetContains(receivers, *to)
}
