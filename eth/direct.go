package eth

import (
	"context"
	"errors"
	"iter"
	"math/big"
	"slices"
	"time"
	"zoro/misc"
	"zoro/rate"
	"zoro/rate/contract"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type directEthClient struct {
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

	dc := &directEthClient{
		chain:    chainId.Text(10),
		client:   ethClient,
		ratefeed: config.RateFeed,
		signer:   types.LatestSignerForChainID(chainId),
	}

	if len(config.Accounts) > 0 {
		dc.accounts.set = misc.SliceToSet(config.Accounts)
		dc.accounts.list = slices.Clone(config.Accounts)
	}

	return dc, nil
}

func (client *directEthClient) RateFeed() common.Address {
	return client.ratefeed
}

func (client *directEthClient) Accounts() iter.Seq[common.Address] {
	return slices.Values(client.accounts.list)
}

func (client *directEthClient) BlockTime() time.Duration {
	return client.blockTime
}

func (client *directEthClient) Chain() string {
	return client.chain
}

func (client *directEthClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return client.client.BalanceAt(ctx, account, nil)
}

func (client *directEthClient) GasPrice(ctx context.Context) (*big.Int, error) {
	return client.client.SuggestGasPrice(ctx)
}

func (client *directEthClient) RateAt(ctx context.Context, block int64) (*rate.Rate, error) {
	out, err := contract.CallGetRate(ctx, client.client, block, client.ratefeed)
	if err != nil {
		return nil, err
	}
	return &rate.Rate{
		Sub: out.Sub,
		Div: out.Div,
		Mul: out.Mul,
		Add: out.Sub,
	}, nil
}

func (client *directEthClient) FinalHeight(ctx context.Context) (int64, error) {
	header, err := client.client.HeaderByNumber(ctx, big.NewInt(FinalBlock))
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}

func (client *directEthClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	return client.client.NonceAt(ctx, account, nil)
}

func (client *directEthClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := client.client.TransactionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if misc.FilterNonReceive(tx, client.chain, client.accounts.set) {
		return nil, false, nil
	}
	return tx, isPending, nil
}

func (client *directEthClient) Submit(ctx context.Context, tx *types.Transaction) error {
	if _, err := client.signer.Sender(tx); err != nil {
		return err
	}
	if misc.FilterNonReceive(tx, client.chain, client.accounts.set) {
		return errors.New("tx is filtered")
	}
	return client.client.SendTransaction(ctx, tx)
}

func (client *directEthClient) Receives(ctx context.Context, number int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	block, err := client.client.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}
	return func(yield func(common.Address, *types.Transaction) bool) {
		for _, tx := range block.Transactions() {
			if misc.FilterNonReceive(tx, client.chain, client.accounts.set) {
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
