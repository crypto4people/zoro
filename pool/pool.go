package pool

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient interface {
	ChainID(ctx context.Context) (*big.Int, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Close()
}

type EthPool struct {
	mu      sync.Mutex
	clients []*ethclient.Client
}

func New(ctx context.Context, nodes []string) (*EthPool, error) {
	pool := &EthPool{}

	for _, node := range nodes {
		client, err := ethclient.Dial(node)
		if err != nil {
			return nil, err
		}
		pool.clients = append(pool.clients, client)
	}

	if len(pool.clients) == 0 {
		return nil, errors.New("no working Ethereum nodes found")
	}

	return pool, nil
}

func (pool *EthPool) ChainID(ctx context.Context) (*big.Int, error) {
	return tryAllClients(pool, func(client *ethclient.Client) (*big.Int, error) {
		return client.ChainID(ctx)
	})
}

func (pool *EthPool) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return tryAllClients(pool, func(client *ethclient.Client) ([]byte, error) {
		return client.CallContract(ctx, msg, blockNumber)
	})
}

func (pool *EthPool) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return tryAllClients(pool, func(client *ethclient.Client) (*types.Header, error) {
		return client.HeaderByNumber(ctx, number)
	})
}

func (pool *EthPool) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return tryAllClients(pool, func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(ctx, number)
	})
}

func (pool *EthPool) Close() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, client := range pool.clients {
		client.Close()
	}
	pool.clients = nil
}

func tryAllClients[R any](pool *EthPool, fn func(*ethclient.Client) (R, error)) (result R, lastErr error) {
	var err error

	for idx, client := range pool.clients {
		if result, err = fn(client); err == nil {
			if idx != 0 {
				pool.mu.Lock()
				pool.clients[0], pool.clients[idx] = pool.clients[idx], pool.clients[0]
				pool.mu.Unlock()
			}
			return result, nil
		} else {
			lastErr = err
		}
	}

	return result, lastErr
}
