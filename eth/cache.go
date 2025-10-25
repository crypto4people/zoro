package eth

import (
	"context"
	"errors"
	"iter"
	"maps"
	"math/big"
	"strconv"
	"time"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/maypok86/otter/v2"
	"golang.org/x/sync/singleflight"
)

var _ Client = &cacheClient{}

type cacheClient struct {
	client Client
	sf     *singleflight.Group
	cache  *otter.Cache[string, any]
}

// Accounts implements Client.
func (client *cacheClient) Accounts() iter.Seq[common.Address] {
	return client.client.Accounts()
}

// BlockTime implements Client.
func (client *cacheClient) BlockTime() time.Duration {
	return client.client.BlockTime()
}

// Chain implements Client.
func (client *cacheClient) Chain() string {
	return client.client.Chain()
}

// RateFeed implements Client.
func (client *cacheClient) RateFeed() common.Address {
	return client.client.RateFeed()
}

// BalanceOf implements Client.
func (client *cacheClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return getOrFetch(client.sf, client.cache, ctx, "BalanceOf:"+account.Hex(), func() (*big.Int, error) {
		return client.client.BalanceOf(ctx, account)
	})
}

// FinalHeight implements Client.
func (client *cacheClient) FinalHeight(ctx context.Context) (int64, error) {
	return getOrFetch(client.sf, client.cache, ctx, "FinalHeight", func() (int64, error) {
		return client.client.FinalHeight(ctx)
	})
}

// GasPrice implements Client.
func (client *cacheClient) GasPrice(ctx context.Context) (*big.Int, error) {
	return getOrFetch(client.sf, client.cache, ctx, "GasPrice", func() (*big.Int, error) {
		return client.client.GasPrice(ctx)
	})
}

// NonceOf implements Client.
func (client *cacheClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	return getOrFetch(client.sf, client.cache, ctx, "NonceOf:"+account.Hex(), func() (uint64, error) {
		return client.client.NonceOf(ctx, account)
	})
}

// RateAt implements Client.
func (client *cacheClient) RateAt(ctx context.Context, block int64) (*rate.Rate, error) {
	return getOrFetch(client.sf, client.cache, ctx, "RateAt:"+strconv.FormatInt(block, 10), func() (*rate.Rate, error) {
		return client.client.RateAt(ctx, block)
	})
}

// Receipt implements Client.
func (client *cacheClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	type receiptResult struct {
		tx        *types.Transaction
		confirmed bool
	}

	result, err := getOrFetch(client.sf, client.cache, ctx, "Receipt:"+hash.Hex(), func() (*receiptResult, error) {
		tx, confirmed, err := client.client.Receipt(ctx, hash)
		if err != nil {
			return nil, err
		}

		return &receiptResult{
			tx:        tx,
			confirmed: confirmed,
		}, nil
	})
	if err != nil {
		return nil, false, err
	}

	return result.tx, result.confirmed, nil
}

// Receives implements Client.
func (client *cacheClient) Receives(ctx context.Context, block int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	mapp, err := getOrFetch(client.sf, client.cache, ctx, "receives:"+strconv.FormatInt(block, 10), func() (map[common.Address]*types.Transaction, error) {
		seq, err := client.client.Receives(ctx, block)
		if err != nil {
			return nil, err
		}

		return maps.Collect(seq), nil
	})
	if err != nil {
		return nil, err
	}

	return maps.All(mapp), nil
}

// Submit implements Client.
func (client *cacheClient) Submit(ctx context.Context, tx *types.Transaction) error {
	return client.client.Submit(ctx, tx)
}

func getOrFetch[T any](sf *singleflight.Group, cache *otter.Cache[string, any], ctx context.Context, key string, fn func() (T, error)) (T, error) {
	var zero T

	resCh := sf.DoChan(key, func() (any, error) {
		if val, ok := cache.GetIfPresent(key); ok {
			return val, nil
		}

		val, err := fn()
		if err != nil {
			return nil, err
		}

		cache.Set(key, val)
		return val, nil
	})

	select {
	case result := <-resCh:
		if result.Err != nil {
			return zero, result.Err
		}

		casted, ok := result.Val.(T)
		if !ok {
			return zero, errors.New("cast failed")
		}
		return casted, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}
