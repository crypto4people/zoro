package hub

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"math/big"
	"slices"
	"time"
	"zoro/eth"
	"zoro/hub/pb"
	"zoro/misc"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ eth.Client = &hubClient{}

type HubClientConfig struct {
	Accounts []common.Address
	URL      string
}

type hubClient struct {
	info     *pb.InfoResult
	client   pb.HubClient
	accounts struct {
		set  map[common.Address]struct{}
		list [][]byte
	}
	signer types.Signer
	chain  string
}

func NewClient(ctx context.Context, cc grpc.ClientConnInterface, config HubClientConfig) (eth.Client, error) {

	client := pb.NewHubClient(cc)

	info, err := client.Info(ctx, nil)

	if err != nil {
		return nil, err
	}

	if len(info.Ratefeed) != 20 {
		return nil, errors.New("invalid ratefeed address length")
	}
	if info.Chain == "" {
		return nil, errors.New("empty chain id")
	}

	if info.Blocktime == nil {
		return nil, errors.New("missing block time")
	}

	if info.Blocktime.AsDuration() <= time.Second {
		return nil, errors.New("invalid block time")
	}

	if info.Height <= 0 {
		return nil, errors.New("invalid height")
	}

	if info.Version == "" {
		return nil, errors.New("empty version")
	}

	chainId, chainIdOK := new(big.Int).SetString(info.Chain, 10)
	if !chainIdOK {
		return nil, fmt.Errorf("failed to parse chain id %q", info.Chain)
	}

	hub := &hubClient{
		info:   info,
		client: client,
		signer: types.LatestSignerForChainID(chainId),
	}

	hub.accounts.set = misc.SliceToSet(config.Accounts)
	hub.accounts.list = lo.Map(config.Accounts, func(addr common.Address, _ int) []byte {
		return slices.Clone(addr[:])
	})

	return hub, nil
}

func (client *hubClient) Accounts() iter.Seq[common.Address] {
	return func(yield func(common.Address) bool) {
		for _, addr := range client.accounts.list {
			if !yield(common.Address(addr)) {
				break
			}
		}
	}
}

func (client *hubClient) Chain() string {
	return client.info.Chain
}

func (client *hubClient) BlockTime() time.Duration {
	return client.info.Blocktime.AsDuration()
}

func (client *hubClient) RateFeed() common.Address {
	return common.BytesToAddress(client.info.Ratefeed)
}

func (client *hubClient) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	res, err := client.client.BalanceOf(ctx, &pb.BalanceOfParams{
		Account: account[:],
		Chain:   client.info.Chain,
	})
	if err != nil {
		return nil, err
	}
	return parseBigInt(res.Balance)
}

func (client *hubClient) FinalHeight(ctx context.Context) (int64, error) {
	info, err := client.client.Info(ctx, nil)
	if err != nil {
		return 0, err
	}
	if info.Chain != client.info.Chain || bytes.Equal(info.Ratefeed, client.info.Ratefeed) {
		return 0, errors.New("inconsistent indexer info")
	}
	if info.Height <= 0 {
		return 0, errors.New("invalid height")
	}
	return info.Height, nil
}

func (client *hubClient) NonceOf(ctx context.Context, account common.Address) (uint64, error) {
	res, err := client.client.NonceOf(ctx, &pb.NonceOfParams{
		Account: account[:],
		Chain:   client.info.Chain,
	})
	if err != nil {
		return 0, err
	}
	return res.Nonce, nil
}

func (client *hubClient) RateAt(ctx context.Context, block int64) (*rate.Rate, error) {
	res, err := client.client.RateAt(ctx, &pb.RateAtParams{
		Block:    block,
		Chain:    client.info.Chain,
		Ratefeed: client.info.Ratefeed,
	})
	if err != nil {
		return nil, err
	}
	return parseRatePb(res)
}

func (client *hubClient) Receipt(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	res, err := client.client.Receipt(ctx, &pb.ReceiptParams{
		Hash: hash[:],
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	tx, err := misc.ParseEthTransaction(res.Tx)
	if err != nil {
		return nil, false, err
	}
	if misc.FilterNonReceive(tx, client.chain, client.accounts.set) {
		return nil, false, nil
	}
	return tx, res.Pend, nil
}

func (client *hubClient) Receives(ctx context.Context, block int64) (iter.Seq2[common.Address, *types.Transaction], error) {
	res, err := client.client.Receives(ctx, &pb.ReceivesParams{
		Block:    block,
		Chain:    client.info.Chain,
		Accounts: client.accounts.list,
	})
	if err != nil {
		return nil, err
	}

	txs := make([]struct {
		tx     types.Transaction
		sender common.Address
	}, len(res.Txs))

	for i := range res.Txs {
		err = txs[i].tx.UnmarshalBinary(res.Txs[i])
		if err != nil {
			return nil, err
		}
		txs[i].sender, err = client.signer.Sender(&txs[i].tx)
		if err != nil {
			return nil, err
		}
	}

	return func(yield func(common.Address, *types.Transaction) bool) {
		for i := range txs {
			if !yield(txs[i].sender, &txs[i].tx) {
				break
			}
		}
	}, nil
}

func (client *hubClient) Submit(ctx context.Context, tx *types.Transaction) (err error) {
	if _, err = client.signer.Sender(tx); err != nil {
		return err
	}
	if misc.FilterNonReceive(tx, client.chain, client.accounts.set) {
		return errors.New("tx is filtered")
	}

	data, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	_, err = client.client.Submit(ctx, &pb.SubmitParams{
		Tx: data,
	})

	return err
}

func parseBigInt(str string) (*big.Int, error) {
	bi, ok := new(big.Int).SetString(str, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse big int %q", str)
	}
	return bi, nil
}

func parseRatePb(res *pb.RateAtResult) (r *rate.Rate, err error) {
	r = &rate.Rate{}

	r.Sub, err = parseBigInt(res.Sub)
	if err != nil {
		return nil, err
	}

	r.Div, err = parseBigInt(res.Div)
	if err != nil {
		return nil, err
	}

	r.Mul, err = parseBigInt(res.Mul)
	if err != nil {
		return nil, err
	}

	r.Add, err = parseBigInt(res.Add)
	if err != nil {
		return nil, err
	}

	return r, nil
}
