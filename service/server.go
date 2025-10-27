package service

import (
	"context"
	"net/http"
	"net/url"
	"zoro/proxy"
	"zoro/service/db"
	"zoro/service/pb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.ServiceServer = &Server{}

type Server struct {
	database   *db.Database
	httpClient *http.Client
	signer     types.Signer
	pb.UnimplementedServiceServer
}

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

		}
	}

	return &pb.CheckTxResult{
		Status: status,
	}, nil

}

func (server *Server) SubmitTx(ctx context.Context, params *pb.SubmitTxParams) (_ *pb.SubmitTxResult, err error) {
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
		err = server.database.PutPend(tx.Hash(), serverURL)
	}

	return &pb.SubmitTxResult{}, err
}

func (server *Server) ListPends(ctx context.Context, params *pb.ListPendsParams) (*pb.ListPendsResult, error) {
	pends, err := server.database.Pends()
	if err != nil {
		return nil, err
	}

	result := &pb.ListPendsResult{
		Txs: make([]*pb.ListPendsResult_Tx, 0, len(pends)),
	}

	for hash, server := range pends {
		result.Txs = append(result.Txs, &pb.ListPendsResult_Tx{
			Hash:   hash.String(),
			Server: server.String(),
		})
	}

	return result, nil
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

	info, err := proxy.NewClient(server.httpClient, params.Url, account).GetInfo(ctx)
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

func (server *Server) ListSubs(ctx context.Context, params *pb.ListSubsParams) (_ *pb.ListSubsResult, err error) {
	address := common.Address{}
	err = address.UnmarshalText([]byte(params.Address))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse address")
	}

	subs, err := proxy.NewClient(server.httpClient, params.Server, address).GetSubs(ctx)
	if err != nil {
		return nil, err
	}

	result := &pb.ListSubsResult{
		Subs: make([]*pb.ListSubsResult_Sub, len(subs)),
	}

	for idx, sub := range subs {
		result.Subs[idx] = &pb.ListSubsResult_Sub{
			Tx:     sub.Tx.Hex(),
			Expire: sub.Expire,
			Remain: sub.Remain,
		}
	}

	return result, nil

}
