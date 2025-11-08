package indexer

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/crypto4people/zoro/hub/indexer"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func runAction(ctx context.Context, cmd *cli.Command) error {
	logger := slog.Default()

	config := cmd.String("config")
	if config == "" {
		return errors.New("config from cmd is not supported yet")
	}

	cfg, err := parseFromFile(config)
	if err != nil {
		return err
	}

	client, err := ethclient.Dial(cfg.EthNode)
	if err != nil {
		return errors.Join(err, errors.New("failed to connect to ethereum node"))
	}
	defer client.Close()

	chain, err := client.ChainID(ctx)
	if err != nil {
		return errors.Join(err, errors.New("failed to get chain id"))
	}

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return err
	}
	defer func() { _ = db.Disconnect(ctx) }()

	err = db.Ping(ctx, nil)
	if err != nil {
		return errors.Join(err, errors.New("failed to connect to mongodb"))
	}

	blockTime, err := time.ParseDuration(cfg.BlockTime)
	if err != nil {
		return errors.Join(err, errors.New("failed to parse block time"))
	}

	indx := indexer.NewIndexer(logger,
		client,
		db.Database(cfg.MongoDatabase, options.Database()),
		chain.String(),
		cfg.RatefeedContract,
		blockTime,
	)

	return indx.Routine(ctx)
}
