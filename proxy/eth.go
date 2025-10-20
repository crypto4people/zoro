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

type EthClient struct {
	client   *ethclient.Client
	contract common.Address
	signer   types.Signer
	chain    string
}

func NewEthClient(ctx context.Context, contract common.Address, node string) (*EthClient, error) {
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

	return &EthClient{
		chain:    chain,
		client:   ethClient,
		contract: contract,
		signer:   signer,
	}, nil
}

func (cli *EthClient) Contract() common.Address {
	return cli.contract
}

func (cli *EthClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return cli.client.BalanceAt(ctx, account, nil)
}

func (cli *EthClient) RateAt(ctx context.Context, block int64) (*Rate, error) {
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

func (cli *EthClient) Chain() string {
	return cli.chain
}

func (cli *EthClient) FinalHeight(ctx context.Context) (int64, error) {
	header, err := cli.client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}

func (cli *EthClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	return cli.client.NonceAt(ctx, account, nil)
}

func (cli *EthClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := cli.client.TransactionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return tx, isPending, nil
}

func (cli *EthClient) Submit(ctx context.Context, tx *types.Transaction) error {
	if _, err := cli.signer.Sender(tx); err != nil {
		return err
	}
	return cli.client.SendTransaction(ctx, tx)
}

func (cli *EthClient) Transactions(ctx context.Context, number int64) (iter.Seq2[common.Address, *types.Transaction], error) {
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

func (cli *EthClient) Shutdown(ctx context.Context) error {
	cli.client.Close()
	return nil
}
