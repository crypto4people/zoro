package eth

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"math/big"
	"slices"
	"time"
	"zoro/hub"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ Client = &hubClient{}

type HubClientConfig struct {
	Accounts []common.Address
	URL      string
}

type hubClient struct {
	info     hub.Info
	client   *hub.Client
	accounts struct {
		set  map[common.Address]struct{}
		list []common.Address
	}
	signer    types.Signer
	chain     string
	blockTime time.Duration
}

func NewHubClient(ctx context.Context, client *hub.Client, config HubClientConfig) (Client, error) {

	info, err := client.Info(ctx)
	if err != nil {
		return nil, err
	}

	chainId, chainIdOK := new(big.Int).SetString(info.Chain, 10)
	if !chainIdOK {
		return nil, fmt.Errorf("failed to parse chain id %q", info.Chain)
	}

	blockTime, err := time.ParseDuration(info.BlockTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block time %q", info.BlockTime)
	}

	signer := types.LatestSignerForChainID(chainId)

	hub := &hubClient{
		info:      *info,
		client:    client,
		signer:    signer,
		blockTime: blockTime,
	}

	hub.accounts.set = sliceToSet(config.Accounts)
	hub.accounts.list = slices.Clone(config.Accounts)

	return hub, nil
}

func (client *hubClient) Accounts() iter.Seq[common.Address] {
	return slices.Values(client.accounts.list)
}

func (client *hubClient) Chain() string {
	return client.info.Chain
}

func (client *hubClient) BlockTime() time.Duration {
	return client.blockTime
}

func (client *hubClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return client.client.BalanceOf(ctx, account, client.info.Chain)
}

func (client *hubClient) FinalHeight(ctx context.Context) (int64, error) {
	info, err := client.client.Info(ctx)
	if err != nil {
		return 0, err
	}
	if info.Chain != client.info.Chain || info.RateFeed != client.info.RateFeed {
		return 0, errors.New("inconsistent indexer info")
	}
	return info.Height, nil
}

func (client *hubClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	return client.client.NonceOf(ctx, account, client.info.Chain)
}

func (client *hubClient) RateAt(ctx context.Context, block int64) (*rate.Rate, error) {
	return client.client.RateAt(ctx, block, client.info.Chain, client.info.RateFeed)
}

func (client *hubClient) RateFeed() common.Address {
	return client.info.RateFeed
}

func (client *hubClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, pend, err := client.client.Receipt(ctx, hash)
	if err != nil {
		return nil, false, err
	}
	if filterReceives(tx, client.chain, client.accounts.set) {
		return nil, false, nil
	}
	return tx, pend, nil
}

func (client *hubClient) Receives(ctx context.Context, block int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	return client.client.Receives(ctx, block, client.info.Chain, client.accounts.list)
}

func (client *hubClient) Submit(ctx context.Context, tx *types.Transaction) error {
	if filterReceives(tx, client.chain, client.accounts.set) {
		return errors.New("tx is filtered")
	}
	return client.client.Submit(ctx, tx)
}
