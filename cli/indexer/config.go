package indexer

import (
	"encoding/json"
	"io"
	"os"

	"github.com/crypto4people/zoro/cli/constants"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v3"
)

type config struct {
	EthNode          string         `json:"eth_node,omitempty"`
	MongoURI         string         `json:"mongo_uri"`
	MongoDatabase    string         `json:"mongo_database,omitempty"`
	RatefeedContract common.Address `json:"ratefeed_contract,omitempty"`
	BlockTime        string         `json:"block_time,omitempty"`
}

func parseFromFile(configFile string) (config, error) {
	var cfg config

	file, err := os.Open(configFile)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	err = json.NewDecoder(io.LimitReader(file, 1<<20)).Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	if cfg.EthNode == "" {
		cfg.EthNode = constants.DEFAULT_ETH_RPC
	}

	if cfg.MongoDatabase == "" {
		cfg.MongoDatabase = constants.MONGO_DATABASE_NAME
	}

	if cfg.RatefeedContract == (common.Address{}) {
		cfg.RatefeedContract = common.HexToAddress(constants.DEFAULT_RATEFEED_CONTRACT_ADDRESS)
	}

	if cfg.BlockTime == "" {
		cfg.BlockTime = constants.ETHEREUM_BLOCK_TIME
	}

	return cfg, nil
}

func parseFromCmd(cmd *cli.Command) (config, error) {
	panic("unimplemented")
}
