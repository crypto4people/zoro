package hub

import (
	"context"
	"time"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

type IndexerInfo struct {
	Uptime    time.Duration
	Version   string
	RateFeed  common.Address
	Chain     string
	Height    int64
	BlockTime time.Duration
}

type Indexer struct {
	client *ethclient.Client
	db     *mongo.Database
}

func (indexer *Indexer) Info(ctx context.Context) (*IndexerInfo, error) {
	panic("todo")
}

func (indexer *Indexer) RateAt(ctx context.Context, chain string, block int64, ratefeed common.Address) (*rate.Rate, error) {
	panic("todo")
}

func (indexer *Indexer) Receipt(ctx context.Context, hash common.Hash) ([]byte, bool, error) {
	panic("todo")
}

func (indexer *Indexer) Receives(ctx context.Context, chain string, block int64, accounts map[common.Address]struct{}) ([][]byte, error) {
	panic("todo")
}

func (indexer *Indexer) Routine(ctx context.Context) error {
	panic("todo")
}
