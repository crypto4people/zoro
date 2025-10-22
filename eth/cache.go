package eth

import (
	"context"
	"iter"
	"math/big"
	"time"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ Client = &cacheClient{}

type cacheClient struct {
	client Client
}

// Accounts implements Client.
func (client *cacheClient) Accounts() iter.Seq[common.Address] {
	panic("unimplemented")
}

// BalanceOf implements Client.
func (client *cacheClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	panic("unimplemented")
}

// BlockTime implements Client.
func (client *cacheClient) BlockTime() time.Duration {
	panic("unimplemented")
}

// Chain implements Client.
func (client *cacheClient) Chain() string {
	panic("unimplemented")
}

// FinalHeight implements Client.
func (client *cacheClient) FinalHeight(ctx context.Context) (int64, error) {
	panic("unimplemented")
}

// GasPrice implements Client.
func (client *cacheClient) GasPrice(ctx context.Context) (*big.Int, error) {
	panic("unimplemented")
}

// NonceOf implements Client.
func (client *cacheClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	panic("unimplemented")
}

// RateAt implements Client.
func (client *cacheClient) RateAt(ctx context.Context, block int64) (*rate.Rate, error) {
	panic("unimplemented")
}

// RateFeed implements Client.
func (client *cacheClient) RateFeed() common.Address {
	panic("unimplemented")
}

// Receipt implements Client.
func (client *cacheClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	panic("unimplemented")
}

// Receives implements Client.
func (client *cacheClient) Receives(ctx context.Context, block int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	panic("unimplemented")
}

// Submit implements Client.
func (client *cacheClient) Submit(ctx context.Context, tx *types.Transaction) error {
	panic("unimplemented")
}
