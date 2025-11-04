package service

import (
	"context"
	"net/http"
	"net/url"

	"github.com/crypto4people/zoro/proxy"
	"github.com/crypto4people/zoro/service/db"
	"github.com/crypto4people/zoro/service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ pb.ServiceServer = &Server{}

type Server struct {
	database   *db.Database
	httpClient *http.Client
	signer     types.Signer
	pb.UnimplementedServiceServer
}

// Connect implements pb.ServiceServer.
func (server *Server) Connect(context.Context, *pb.ConnectRequest) (*pb.ConnectResult, error) {
	panic("unimplemented")
}

// Disconnect implements pb.ServiceServer.
func (server *Server) Disconnect(context.Context, *pb.DisconnectRequest) (*pb.DisconnectResult, error) {
	panic("unimplemented")
}

// Stats implements pb.ServiceServer.
func (server *Server) Stats(context.Context, *pb.StatsRequest) (*pb.StatsResult, error) {
	panic("unimplemented")
}

// CheckTx implements pb.ServiceServer.
func (server *Server) CheckTx(ctx context.Context, params *pb.CheckTxParams) (*pb.CheckTxResult, error) {
	serverURL, err := url.Parse(params.Server)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse server url")
	}

	hash := common.Hash{}
	err = hash.UnmarshalText([]byte(params.Hash))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse hash")
	}

	tx, isPend, err := proxy.NewClient(server.httpClient, serverURL.String(), common.Address{}).GetReceipt(ctx, hash)

	if err != nil {
		return nil, err
	}
	var status pb.CheckTxResult_TxStatus
	if tx == nil {
		status = pb.CheckTxResult_NotFound
		_ = server.database.PutPend(tx.Hash(), nil)
	} else {
		if isPend {
			status = pb.CheckTxResult_Pend
			sender, err := server.signer.Sender(tx)
			if err == nil {
				has, err := server.database.HasAccount(sender)
				if err == nil {
					if has {
						_ = server.database.PutPend(tx.Hash(), serverURL)
					}
				}
			}
		} else {
			status = pb.CheckTxResult_Done
			_ = server.database.PutPend(tx.Hash(), nil)
		}
	}

	return &pb.CheckTxResult{
		Status: status,
	}, nil
}

// GetAccount implements pb.ServiceServer.
func (server *Server) GetAccount(ctx context.Context, params *pb.GetAccountParams) (*pb.GetAccountResult, error) {
	address := common.Address{}
	err := address.UnmarshalText([]byte(params.Account))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse hash")
	}

	cipher, err := server.database.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccountResult{
		Cipher: cipher,
	}, nil
}

// ListAccounts implements pb.ServiceServer.
func (server *Server) ListAccounts(context.Context, *pb.ListAccountsParams) (*pb.ListAccountsResult, error) {
	accounts, err := server.database.ListAccounts()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, account := range accounts {
		result = append(result, account.Hex())
	}

	return &pb.ListAccountsResult{
		Accounts: result,
	}, nil
}

// ListPends implements pb.ServiceServer.
func (server *Server) ListPends(context.Context, *pb.ListPendsParams) (*pb.ListPendsResult, error) {
	pends, err := server.database.Pends()
	if err != nil {
		return nil, err
	}

	result := make([]*pb.ListPendsResult_Tx, 0, len(pends))
	for hash, serverURL := range pends {
		result = append(result, &pb.ListPendsResult_Tx{
			Hash:   hash.Hex(),
			Server: serverURL.String(),
		})
	}

	return &pb.ListPendsResult{
		Txs: result,
	}, nil

}

// ListServers implements pb.ServiceServer.
func (server *Server) ListServers(context.Context, *pb.ListServersParams) (*pb.ListServersResult, error) {
	servers, err := server.database.ListServers()
	if err != nil {
		return nil, err
	}

	result := make([]*pb.ListServersResult_Server, 0, len(servers))
	for name, url := range servers {
		result = append(result, &pb.ListServersResult_Server{
			Key: name,
			Url: url,
		})
	}

	return &pb.ListServersResult{
		Servers: result,
	}, nil

}

// ListSubs implements pb.ServiceServer.
func (server *Server) ListSubs(ctx context.Context, params *pb.ListSubsParams) (*pb.ListSubsResult, error) {
	serverURL, err := url.Parse(params.Server)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse server url")
	}

	client := proxy.NewClient(server.httpClient, serverURL.String(), common.HexToAddress(params.Address))

	subs, err := client.GetSubs(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.ListSubsResult_Sub, 0, len(subs))
	for _, sub := range subs {
		result = append(result, &pb.ListSubsResult_Sub{
			Tx:     sub.Tx.Hex(),
			Expire: sub.Expire,
			Remain: sub.Remain,
		})
	}

	return &pb.ListSubsResult{
		Subs: result,
	}, nil
}

// ModifyAccount implements pb.ServiceServer.
func (server *Server) ModifyAccount(ctx context.Context, params *pb.ModifyAccountParams) (*pb.ModifyAccountResult, error) {
	address := common.Address{}
	err := address.UnmarshalText([]byte(params.Address))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse hash")
	}

	err = server.database.SaveAccount(address, params.Cipher)
	if err != nil {
		return nil, err
	}

	return &pb.ModifyAccountResult{}, nil
}

// ModifyServer implements pb.ServiceServer.
func (server *Server) ModifyServer(ctx context.Context, params *pb.ModifyServerParams) (*pb.ModifyServerResult, error) {
	serverURL, err := url.Parse(params.Url)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse server url")
	}

	err = server.database.SaveServer(params.Key, serverURL.String())
	if err != nil {
		return nil, err
	}

	return &pb.ModifyServerResult{}, nil
}

// ServerInfo implements pb.ServiceServer.
func (server *Server) ServerInfo(ctx context.Context, params *pb.ServerInfoParams) (*pb.ServerInfoResult, error) {
	serverURL, err := url.Parse(params.Url)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse server url")
	}

	client := proxy.NewClient(server.httpClient, serverURL.String(), common.HexToAddress(params.Account))

	info, err := client.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ServerInfoResult{
		Version:  info.Version,
		Uptime:   info.Uptime,
		Conns:    info.Conns,
		Subs:     info.Subs,
		Chain:    info.Chain.Dec(),
		Height:   info.Height,
		RateFeed: info.RateFeed.Hex(),
		Account:  info.Account.Hex(),
		GasPrice: info.GasPrice.Dec(),
		Rate: &pb.ServerInfoResult_Rate{
			Sub: info.Rate.Sub.Dec(),
			Div: info.Rate.Div.Dec(),
			Mul: info.Rate.Mul.Dec(),
			Add: info.Rate.Add.Dec(),
		},
	}, nil
}

// SubmitTx implements pb.ServiceServer.
func (server *Server) SubmitTx(ctx context.Context, params *pb.SubmitTxParams) (*pb.SubmitTxResult, error) {
	serverURL, err := url.Parse(params.Server)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse server url")
	}

	tx := &types.Transaction{}
	err = tx.UnmarshalBinary(params.Tx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse tx")
	}

	sender, err := server.signer.Sender(tx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to get tx signer")
	}

	has, err := server.database.HasAccount(sender)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, status.Error(codes.NotFound, "failed to get tx signer")
	}

	err = proxy.NewClient(server.httpClient, serverURL.String(), common.Address{}).SubmitTx(ctx, tx)

	if err == nil {
		_ = server.database.PutPend(tx.Hash(), serverURL)
	}

	return &pb.SubmitTxResult{}, err
}
