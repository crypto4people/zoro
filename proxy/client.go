package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/ethereum/go-ethereum/common"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
	account    common.Address
}

func NewClient(client *http.Client, endpoint string, account common.Address) *Client {
	return &Client{
		httpClient: client,
		endpoint:   endpoint,
		account:    account,
	}
}

func (client *Client) Endpoint() string {
	return client.endpoint
}

func (client *Client) Account() common.Address {
	return client.account
}

func (client *Client) GetInfo(ctx context.Context) (*Info, error) {
	return doRequest[Info](ctx, client.httpClient, client.endpoint)
}

func (client *Client) GetSubs(ctx context.Context) ([]Sub, error) {
	url, err := url.Parse(client.endpoint)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "subs")

	query := url.Query()
	query.Set("user", client.account.Hex())
	url.RawQuery = query.Encode()

	result, err := doRequest[[]Sub](ctx, client.httpClient, url.String())
	if err != nil || result == nil {
		return nil, err
	}

	return *result, nil
}

func doRequest[R any](ctx context.Context, client *http.Client, endpoint string) (*R, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result R
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
