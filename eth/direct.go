package eth

import (
	"context"
	"errors"
	"iter"
	"math/big"
	"slices"
	"time"
	"zoro/contract"
	"zoro/rate"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type directClient struct {
	client   *ethclient.Client
	ratefeed common.Address
	accounts struct {
		set  map[common.Address]struct{}
		list []common.Address
	}
	signer    types.Signer
	chain     string
	blockTime time.Duration
}

type DirectClientConfig struct {
	Accounts  []common.Address
	RateFeed  common.Address
	BlockTime time.Duration
}

func NewDirectClient(ctx context.Context, ethClient *ethclient.Client, config DirectClientConfig) (Client, error) {

	chainId, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	_, err = contract.CallGetRate(ctx, ethClient, rpc.FinalizedBlockNumber.Int64(), config.RateFeed)
	if err != nil {
		return nil, err
	}

	signer := types.LatestSignerForChainID(chainId)
	chain := chainId.Text(10)

	dc := &directClient{
		chain:    chain,
		client:   ethClient,
		ratefeed: config.RateFeed,
		signer:   signer,
	}

	dc.accounts.set = sliceToSet(config.Accounts)
	dc.accounts.list = slices.Clone(config.Accounts)

	return dc, nil
}

func (client *directClient) RateFeed() common.Address {
	return client.ratefeed
}

func (client *directClient) Accounts() iter.Seq[common.Address] {
	return slices.Values(client.accounts.list)
}

func (client *directClient) BlockTime() time.Duration {
	return client.blockTime
}

func (client *directClient) Chain() string {
	return client.chain
}

func (client *directClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return client.client.BalanceAt(ctx, account, nil)
}

func (client *directClient) RateAt(ctx context.Context, block int64) (*rate.Rate, error) {
	out, err := contract.CallGetRate(ctx, client.client, block, client.ratefeed)
	if err != nil {
		return nil, err
	}
	rate := rate.FromGetRateOutput(*out)
	return &rate, nil
}

func (client *directClient) FinalHeight(ctx context.Context) (int64, error) {
	header, err := client.client.HeaderByNumber(ctx, big.NewInt(FinalBlock))
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}

func (client *directClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	return client.client.NonceAt(ctx, account, nil)
}

func (client *directClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := client.client.TransactionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if filterReceives(tx, client.chain, client.accounts.set) {
		return nil, false, nil
	}
	return tx, isPending, nil
}

func (client *directClient) Submit(ctx context.Context, tx *types.Transaction) error {
	if _, err := client.signer.Sender(tx); err != nil {
		return err
	}
	if filterReceives(tx, client.chain, client.accounts.set) {
		return errors.New("tx is filtered")
	}
	return client.client.SendTransaction(ctx, tx)
}

func (client *directClient) Receives(ctx context.Context, number int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	block, err := client.client.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}
	return func(yield func(common.Address, *types.Transaction) bool) {
		for _, tx := range block.Transactions() {
			if filterReceives(tx, client.chain, client.accounts.set) {
				continue
			}
			sender, err := client.signer.Sender(tx)
			if err != nil {
				panic(err)
			}
			if !yield(sender, tx) {
				break
			}
		}
	}, nil
}
