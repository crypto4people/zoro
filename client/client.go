package client

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Client struct {
	client   *http.Client
	endpoint string
	account  common.Address
}

func New(endpoint string, account common.Address) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		endpoint: endpoint,
		account:  account,
	}
}

type Info struct {
	Version string
	Uptime  int
	Conns   int
	Subs    int
	Chain   *big.Int
	Height  int64
	Account common.Address
}

func (cli *Client) GetInfo(ctx context.Context) (*Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := cli.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info Info
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}
