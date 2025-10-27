package service

import (
	"context"
	"net/http"
	"net/url"
	"zoro/proxy"
	"zoro/service/db"
	"zoro/service/pb"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.ServiceServer = &Server{}

type Server struct {
	database   *db.Database
	httpClient *http.Client
	pb.UnimplementedServiceServer
}

func (server *Server) ListAccounts(ctx context.Context, params *pb.ListAccountsParams) (_ *pb.ListAccountsResult, err error) {
	accounts, err := server.database.Accounts()
	if err != nil {
		return nil, err
	}

	result := &pb.ListAccountsResult{
		Accounts: make([]*pb.ListAccountsResult_Account, 0, len(accounts)),
	}

	for addr, cipher := range accounts {
		if !params.Ciphers {
			cipher = nil
		}
		result.Accounts = append(result.Accounts, &pb.ListAccountsResult_Account{
			Address: addr.String(),
			Cipher:  cipher,
		})
	}

	return result, nil
}

func (server *Server) ModifyAccount(ctx context.Context, params *pb.ModifyAccountParams) (_ *pb.ModifyAccountResult, err error) {
	var addr common.Address
	err = addr.UnmarshalText([]byte(params.Address))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse address")
	}
	if len(params.Cipher) == 0 {
		err = server.database.DeleteAccount(addr)
	} else {
		err = server.database.PutAccount(addr, params.Cipher)
	}
	return &pb.ModifyAccountResult{}, err
}

func (server *Server) ListServers(ctx context.Context, params *pb.ListServersParams) (*pb.ListServersResult, error) {
	servers, err := server.database.Servers()
	if err != nil {
		return nil, err
	}

	result := &pb.ListServersResult{
		Servers: make([]*pb.ListServersResult_Server, 0, len(servers)),
	}

	for key, url := range servers {
		result.Servers = append(result.Servers, &pb.ListServersResult_Server{
			Key: key,
			Url: url.String(),
		})
	}

	return result, nil

}

func (server *Server) ModifyServer(ctx context.Context, params *pb.ModifyServerParams) (_ *pb.ModifyServerResult, err error) {
	if params.Url == nil {
		err = server.database.DeleteServer(params.Key)
	} else {
		var purl *url.URL
		purl, err = url.Parse(*params.Url)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to parse url")
		}
		err = server.database.PutServer(params.Key, purl)
	}
	return &pb.ModifyServerResult{}, err
}

func (server *Server) ServerInfo(ctx context.Context, params *pb.ServerInfoParams) (_ *pb.ServerInfoResult, err error) {
	account := common.Address{}
	if params.Account != nil {
		err = account.UnmarshalText([]byte(*params.Account))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to parse account")
		}
	}

	client := proxy.NewClient(server.httpClient, params.Url, account)
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
		Balance:  info.Balance.Dec(),
		GasPrice: info.GasPrice.Dec(),
		Rate: &pb.ServerInfoResult_Rate{
			Sub: info.Rate.Sub.Dec(),
			Div: info.Rate.Div.Dec(),
			Mul: info.Rate.Mul.Dec(),
			Add: info.Rate.Add.Dec(),
		},
	}, nil

}
