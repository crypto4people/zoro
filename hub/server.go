package hub

import (
	"context"
	"time"
	"zoro/eth"
	"zoro/hub/indexer"
	"zoro/hub/pb"
	"zoro/misc"

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
	indexer *indexer.Indexer
	pb.UnimplementedHubServer
}

func NewServer(client eth.Client, indexer *indexer.Indexer) pb.HubServer {
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
		Balance: bal.Text(10),
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
	return nil, ser.client.Submit(ctx, tx)
}

func (ser *server) Info(ctx context.Context, _ *emptypb.Empty) (*pb.InfoResult, error) {
	info, err := ser.indexer.Info(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.InfoResult{
		Version:   version,
		Uptime:    durationpb.New(time.Since(ser.uptime)),
		Height:    info.Height,
		Chain:     info.Chain,
		Ratefeed:  info.Ratefeed[:],
		Blocktime: durationpb.New(info.BlockTime),
	}, nil
}

func (ser *server) RateAt(ctx context.Context, params *pb.RateAtParams) (*pb.RateAtResult, error) {
	res, err := ser.indexer.RateAt(ctx, params.Chain, params.Block, common.BytesToAddress(params.Ratefeed))
	if err != nil {
		return nil, err
	}
	return &pb.RateAtResult{
		Sub: res.Sub.Text(10),
		Div: res.Div.Text(10),
		Mul: res.Mul.Text(10),
		Add: res.Add.Text(10),
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
		return status.Error(codes.NotFound, "tx nto found")
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
