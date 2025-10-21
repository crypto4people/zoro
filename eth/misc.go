package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func filterReceives(tx *types.Transaction, chain string, receivers map[common.Address]struct{}) bool {
	to := tx.To()
	return to == nil || tx.ChainId().Text(10) != chain || tx.Value().Sign() <= 0 || len(tx.Data()) > 0 || !setContaions(receivers, *to)
}

func setContaions[T comparable](mapp map[T]struct{}, v T) bool {
	_, exists := mapp[v]
	return exists
}

func sliceToSet[T comparable](slice []T) map[T]struct{} {
	mapp := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		mapp[v] = struct{}{}
	}
	return mapp
}
