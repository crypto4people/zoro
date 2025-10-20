package proxy

import (
	"context"
	"errors"
	"iter"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient interface {
	RateAt(ctx context.Context, block int64) (*Rate, error)
	Transactions(ctx context.Context, block int64) (iter.Seq2[common.Address, *types.Transaction], error)
	FinalHeight(ctx context.Context) (int64, error)
	NonceOf(ctx context.Context, account common.Address) (uint64, error)
	BalanceOf(ctx context.Context, account common.Address) (*big.Int, error)
	Submit(ctx context.Context, tx *types.Transaction) error
	Receipt(ctx context.Context, hash common.Hash) (tx *types.Transaction, pend bool, err error)
	Shutdown(ctx context.Context) error
	Contract() common.Address
	Chain() string
}

var _ EthClient = &directEthClient{}

type directEthClient struct {
	client   *ethclient.Client
	contract common.Address
	signer   types.Signer
	chain    string
}

func (cli *directEthClient) Contract() common.Address {
	return cli.contract
}

func (cli *directEthClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return cli.client.BalanceAt(ctx, account, nil)
}

func (cli *directEthClient) RateAt(ctx context.Context, block int64) (*Rate, error) {
	result, err := cli.client.CallContract(ctx, ethereum.CallMsg{
		To:   &cli.contract,
		Data: abi.PackGetRate(),
	}, big.NewInt(block))
	if err != nil {
		return nil, err
	}

	rate, err := abi.UnpackGetRate(result)
	if err != nil {
		return nil, err
	}

	return &Rate{
		Sub: rate.Sub,
		Div: rate.Div,
		Mul: rate.Mul,
		Add: rate.Sub,
	}, nil
}

func (cli *directEthClient) Chain() string {
	return cli.chain
}

func (cli *directEthClient) FinalHeight(ctx context.Context) (int64, error) {
	header, err := cli.client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}

func (cli *directEthClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	return cli.client.NonceAt(ctx, account, nil)
}

func (cli *directEthClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := cli.client.TransactionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return tx, isPending, nil
}

func (cli *directEthClient) Submit(ctx context.Context, tx *types.Transaction) error {
	return cli.client.SendTransaction(ctx, tx)
}

func (cli *directEthClient) Transactions(ctx context.Context, number int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	block, err := cli.client.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}
	return func(yield func(common.Address, *types.Transaction) bool) {
		for _, tx := range block.Transactions() {
			sender, err := cli.signer.Sender(tx)
			if err != nil {
				panic(err)
			}
			if !yield(sender, tx) {
				break
			}
		}
	}, nil
}

func (cli *directEthClient) Shutdown(ctx context.Context) error {
	cli.client.Close()
	return nil
}

func DialEthClient(ctx context.Context, contract common.Address, node string) (EthClient, error) {
	ethClient, err := ethclient.DialContext(ctx, node)
	if err != nil {
		return nil, err
	}

	keep := false

	defer func() {
		if !keep {
			ethClient.Close()
		}
	}()

	chainId, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	signer := types.LatestSignerForChainID(chainId)
	chain := chainId.Text(10)

	keep = true

	return &directEthClient{
		chain:    chain,
		client:   ethClient,
		contract: contract,
		signer:   signer,
	}, nil
}
