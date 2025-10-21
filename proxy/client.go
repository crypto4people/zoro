package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
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
	url, err := url.Parse(client.endpoint)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(url.Path, "info")

	query := url.Query()
	query.Set("user", client.account.Hex())
	url.RawQuery = query.Encode()

	return doJsonRequest[Info](ctx, client.httpClient, url.String(), http.NoBody)
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

	result, err := doJsonRequest[[]Sub](ctx, client.httpClient, url.String(), http.NoBody)
	if result == nil {
		return nil, err
	}

	return *result, nil
}

func (client *Client) GetBalance(ctx context.Context) (*uint256.Int, error) {
	zero := uint256.NewInt(0)

	url, err := url.Parse(client.endpoint)
	if err != nil {
		return zero, err
	}

	url.Path = path.Join(url.Path, "eth/balance")

	query := url.Query()
	query.Set("address", client.account.Hex())
	url.RawQuery = query.Encode()

	result, err := doPlainRequest(ctx, client.httpClient, url.String(), http.NoBody)
	if err != nil || result == "" {
		return zero, err
	}

	return uint256.MustFromDecimal(result), nil
}

func (client *Client) GetNonce(ctx context.Context) (uint64, error) {
	zero := uint64(0)

	url, err := url.Parse(client.endpoint)
	if err != nil {
		return zero, err
	}

	url.Path = path.Join(url.Path, "eth/nonce")

	query := url.Query()
	query.Set("address", client.account.Hex())
	url.RawQuery = query.Encode()

	result, err := doPlainRequest(ctx, client.httpClient, url.String(), http.NoBody)
	if err != nil || result == "" {
		return zero, err
	}

	return strconv.ParseUint(result, 10, 32)
}

func (client *Client) GetReceipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	url, err := url.Parse(client.endpoint)
	if err != nil {
		return nil, false, err
	}

	url.Path = path.Join(url.Path, "eth/receipt")

	query := url.Query()
	query.Set("hash", hash.Hex())
	url.RawQuery = query.Encode()

	result, err := doJsonRequest[types.Transaction](ctx, client.httpClient, url.String(), http.NoBody)
	if err != nil {
		return nil, false, err
	}

	//todo fix get reicept by status code
	return result, false, nil
}

func (client *Client) SubmitTx(ctx context.Context, tx *types.Transaction) error {

	url, err := url.Parse(client.endpoint)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, "eth/submit")

	jsonTx, err := tx.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = doPlainRequest(ctx, client.httpClient, url.String(), bytes.NewReader(jsonTx))

	return err
}

func doPlainRequest(ctx context.Context, client *http.Client, endpoint string, body io.Reader) (string, error) {
	respBody, err := doRequest(ctx, client, endpoint, body, func(ctx context.Context, reader io.Reader) ([]byte, error) {
		return io.ReadAll(reader)
	})
	if err != nil || body == nil {
		return "", err
	}

	return string(respBody), nil
}

func doJsonRequest[R any](ctx context.Context, client *http.Client, endpoint string, body io.Reader) (*R, error) {
	return doRequest(ctx, client, endpoint, body, func(ctx context.Context, reader io.Reader) (*R, error) {
		var result R
		err := json.NewDecoder(reader).Decode(&result)
		if err != nil {
			return nil, err
		}

		return &result, nil
	})
}

func doRequest[R any](ctx context.Context, client *http.Client, endpoint string, body io.Reader, fn func(context.Context, io.Reader) (R, error)) (R, error) {
	var zero R

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, body)
	if err != nil {
		return zero, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return zero, errors.New("http error") // todo, send the error back
	}
	return fn(ctx, io.LimitReader(resp.Body, 4<<20))
}
