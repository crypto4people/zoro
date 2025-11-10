package indexer

import (
	"context"
	"time"

	"github.com/crypto4people/zoro/eth"
	"github.com/crypto4people/zoro/indexer/pb"
	"github.com/crypto4people/zoro/misc"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

const version = "beta"

var _ pb.HubServer = &server{}

type server struct {
	uptime  time.Time
	client  eth.Client
	indexer *Indexer
	pb.UnimplementedHubServer
}

func NewServer(client eth.Client, indexer *Indexer) pb.HubServer {
	return &server{
		uptime:  time.Now(),
		client:  client,
		indexer: indexer,
	}
}

func (ser *server) BalanceOf(ctx context.Context, params *pb.BalanceOfParams) (*pb.BalanceOfResult, error) {
	bal, err := ser.client.BalanceOf(ctx, common.BytesToAddress(params.Account))
	if err != nil {
		return nil, err
	}
	return &pb.BalanceOfResult{
		Balance: bal.Dec(),
	}, nil
}

func (ser *server) NonceOf(ctx context.Context, params *pb.NonceOfParams) (*pb.NonceOfResult, error) {
	non, err := ser.client.NonceOf(ctx, common.BytesToAddress(params.Account))
	if err != nil {
		return nil, err
	}
	return &pb.NonceOfResult{
		Nonce: non,
	}, nil
}

func (ser *server) Submit(ctx context.Context, params *pb.SubmitParams) (*emptypb.Empty, error) {
	tx, err := misc.ParseEthTransaction(params.Tx)
	if err != nil {
		return nil, err
	}
	txBytes, err := ser.indexer.GetTx(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}
	if len(txBytes) > 0 {
		return nil, nil
	}
	return nil, ser.client.Submit(ctx, tx)
}

func (ser *server) Info(ctx context.Context, _ *emptypb.Empty) (*pb.InfoResult, error) {

	gasPrice, err := ser.client.GasPrice(ctx)
	if err != nil {
		return nil, err
	}

	info, err := ser.indexer.Info(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.InfoResult{
		Version:   version,
		Uptime:    durationpb.New(time.Since(ser.uptime)),
		Height:    info.Height,
		Chain:     info.Chain,
		RateFeed:  info.Ratefeed[:],
		BlockTime: durationpb.New(info.BlockTime),
		GasPrice:  gasPrice.Dec(),
	}, nil
}

func (ser *server) RateAt(ctx context.Context, params *pb.RateAtParams) (*pb.RateAtResult, error) {
	res, err := ser.indexer.RateAt(ctx, params.Chain, params.Block, common.BytesToAddress(params.Ratefeed))
	if err != nil {
		return nil, err
	}
	return &pb.RateAtResult{
		Sub: res.Sub.Dec(),
		Div: res.Div.Dec(),
		Mul: res.Mul.Dec(),
		Add: res.Add.Dec(),
	}, nil
}

func (ser *server) Receipt(ctx context.Context, params *pb.ReceiptParams) (*pb.ReceiptResult, error) {
	hash := common.BytesToHash(params.Hash)
	txBytes, err := ser.indexer.GetTx(ctx, hash)
	if err != nil {
		return nil, err
	}
	if len(txBytes) > 0 {
		return &pb.ReceiptResult{
			Tx:   txBytes,
			Pend: false,
		}, nil
	}
	tx, pend, err := ser.client.Receipt(ctx, hash)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, status.Error(codes.NotFound, "tx not found")
	}
	txBytes, err = tx.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return &pb.ReceiptResult{
		Tx:   txBytes,
		Pend: pend,
	}, nil

}

func (ser *server) Receives(ctx context.Context, params *pb.ReceivesParams) (*pb.ReceivesResult, error) {
	accounts := make(map[common.Address]struct{}, len(params.Accounts))
	for _, addr := range params.Accounts {
		accounts[common.BytesToAddress(addr)] = struct{}{}
	}
	txs, err := ser.indexer.Receives(ctx, params.Chain, params.Block, accounts)
	if err != nil {
		return nil, err
	}
	return &pb.ReceivesResult{
		Txs: txs,
	}, nil
}
