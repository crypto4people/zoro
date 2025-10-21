package hub

import (
	"context"
	"iter"
	"math/big"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Info struct {
	Version   string
	Height    int64
	Uptime    int64
	Chain     string
	RateFeed  common.Address
	BlockTime string
}

func NewClient(url string) *Client {
	return nil
}

type Client struct {
}

func (client *Client) BalanceOf(ctx context.Context, account common.Address, chain string) (*big.Int, error) {
	panic("unimplemented")
}

func (client *Client) Info(ctx context.Context) (*Info, error) {
	panic("unimplemented")
}

func (client *Client) NonceOf(ctx context.Context, account common.Address, chain string) (uint64, error) {
	panic("unimplemented")
}

func (client *Client) RateAt(ctx context.Context, block int64, chain string, ratefeed common.Address) (*rate.Rate, error) {
	panic("unimplemented")
}

func (client *Client) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	panic("unimplemented")
}

func (client *Client) Receives(ctx context.Context, block int64, chain string, accounts []common.Address) (iter.Seq2[common.Address, *types.Transaction], error) {
	panic("unimplemented")
}

func (client *Client) Submit(ctx context.Context, tx *types.Transaction) error {
	panic("unimplemented")
}
