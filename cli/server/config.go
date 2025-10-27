package server

import (
	"encoding/json"
	"io"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v3"
)

const DEFAULT_ETH_RPC = "https://ethereum-rpc.publicnode.com"
const DEFAULT_RATEFEED_CONTRACT_ADDRESS = "0x2dAD6cBd1B21F99d74409CA24ef53dB6F9789d5E"

type config struct {
	Listen        string           `json:"listen,omitempty"`
	Domain        string           `json:"domain,omitempty"`
	Source        string           `json:"source,omitempty"`
	EthNode       string           `json:"eth_node,omitempty"`
	Accounts      []common.Address `json:"accounts"`
	Contract      common.Address   `json:"contract,omitempty"`
	DataDir       string           `json:"datadir,omitempty"`
	StartingBlock int64            `json:"starting_block,omitempty"`
	source        net.IP
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

	if cfg.Listen == "" {
		cfg.Listen = ":2040"
	}

	if cfg.EthNode == "" {
		cfg.EthNode = DEFAULT_ETH_RPC
	}

	if cfg.Contract == (common.Address{}) {
		cfg.Contract = common.HexToAddress(DEFAULT_RATEFEED_CONTRACT_ADDRESS)
	}

	if cfg.DataDir == "" {
		cfg.DataDir = "./zoro"
	}

	return cfg, nil
}

func parseFromCmd(cmd *cli.Command) (config, error) {
	var cfg config
	return cfg, nil
}
