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

type Client interface {
	BlockTime() time.Duration
	Chain() string
	RateFeed() common.Address
	Accounts() iter.Seq[common.Address]
	BalanceOf(ctx context.Context, account common.Address) (*big.Int, error)
	RateAt(ctx context.Context, block int64) (*rate.Rate, error)
	FinalHeight(ctx context.Context) (int64, error)
	NonceOf(ctx context.Context, account common.Address) (uint64, error)
	Submit(ctx context.Context, tx *types.Transaction) error
	Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	GasPrice(ctx context.Context) (*big.Int, error)
	Receives(ctx context.Context, block int64) (iter.Seq2[common.Address, *types.Transaction], error)
}
