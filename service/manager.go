package service

import (
	"context"
	"math/big"
	"net/http"
	"zoro/proxy"
	"zoro/service/pb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.ServiceServer = &Manager{}

type Manager struct {
	storage    *Storage
	keyring    *Keyring
	address    common.Address
	ethClient  *ethclient.Client
	chain      *big.Int
	httpClient *http.Client
	contract   common.Address

	pb.UnimplementedServiceServer
}

func (man *Manager) Connect(ctx context.Context, params *pb.ConnectParams) (*pb.ConnectResult, error) {
	panic("unimplemented")
}

func (man *Manager) Disconnect(context.Context, *pb.DisconnectParams) (*pb.DisconnectResult, error) {
	panic("unimplemented")
}

func (man *Manager) DeleteServer(ctx context.Context, params *pb.DeleteServerParams) (*pb.DeleteServerResult, error) {
	return nil, man.storage.DeleteServers(ctx, params.Server)
}

func (man *Manager) ListPendingTxs(ctx context.Context, _ *pb.ListPendingTxsParams) (*pb.ListPendingTxsResult, error) {
	return &pb.ListPendingTxsResult{
		Txs: man.storage.ListPendingTxs(ctx),
	}, nil
}

func (man *Manager) ListServers(ctx context.Context, _ *pb.ListServersParams) (*pb.ListServersResult, error) {
	return &pb.ListServersResult{
		Servers: man.storage.ListServers(ctx),
	}, nil
}

func (man *Manager) GetServerSubs(ctx context.Context, params *pb.GetServerSubsParams) (*pb.GetServerSubsResult, error) {
	client := proxy.NewClient(man.httpClient, params.Server, man.address)
	subs, err := client.GetSubs(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetServerSubsResult{
		List: lo.Map(subs, func(sub proxy.Sub, _ int) *pb.GetServerSubsResult_ServerSub {
			return &pb.GetServerSubsResult_ServerSub{
				Expire: &timestamppb.Timestamp{Seconds: int64(sub.Expire)},
				Remain: &durationpb.Duration{Seconds: int64(sub.Remain)},
			}
		}),
	}, nil
}

func (man *Manager) GetServerInfo(ctx context.Context, params *pb.GetServerInfoParams) (*pb.GetServerInfoResult, error) {
	client := proxy.NewClient(man.httpClient, params.Server, man.address)
	info, err := client.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if info.Chain.Cmp(man.chain) == 0 && info.Contract == man.contract {
		err = man.storage.AddServers(ctx, params.Server)
		if err != nil {
			return nil, err
		}
	}
	return &pb.GetServerInfoResult{
		Version: info.Version,
		Uptime:  &durationpb.Duration{Seconds: int64(info.Uptime)},
		Conns:   int64(info.Conns),
		Subs:    int64(info.Subs),
		Chain:   info.Chain.String(),
		Height:  info.Height,
		Address: info.Account.Hex(),
	}, nil
}

func (man *Manager) Purchase(context.Context, *pb.PurchaseParams) (*pb.PurchaseResult, error) {
	panic("unimplemented")
}

func (man *Manager) Status(context.Context, *pb.StatusParams) (*pb.StatusResult, error) {

}
