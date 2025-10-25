package service

import (
	"context"
	"errors"
	"net/http"
	"time"
	"zoro/proxy"
	"zoro/service/db"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
)

type Manager struct {
	repo       *repo
	chain      string
	account    common.Address
	httpClient *http.Client
}

type ConnectParams struct {
	Server string
}

func (man *Manager) Connect(ctx context.Context, params *ConnectParams) error {
	panic("todo")

}

func (man *Manager) Disconnect(ctx context.Context) error {
	panic("todo")

}

type Status struct {
	Version   string
	Uptime    time.Duration
	Account   common.Address
	Chain     string
	Connected bool
}

func (man *Manager) Status(ctx context.Context) (*Status, error) {
	panic("todo")
}

type ModifyServerParams struct {
	Server string
	Add    bool
}

func (man *Manager) ModifyServer(ctx context.Context, params *ModifyServerParams) (err error) {
	if params.Add {
		server := &db.Server{
			URL: params.Server,
		}
		err = man.loadServerInfo(ctx, server)
		if err != nil {
			return err
		}
		return man.repo.UpsertServer(server)
	} else {
		return man.repo.RemoveServer(params.Server)
	}
}

type ListServersResult struct {
	List []db.Server
}

func (man *Manager) ListServers(ctx context.Context) (*ListServersResult, error) {
	list, err := man.repo.ListServers()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	for i := range list {
		if now.Sub(list[i].Time) > time.Second*10 {
			err = man.loadServerInfo(ctx, &list[i])
			if err == nil {
				err = man.repo.UpsertServer(&list[i])
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &ListServersResult{
		List: list,
	}, nil

}

type ListTsxParams struct {
	Cont uint64
}

type ListTxsResult struct {
	List []db.Tx
}

func (man *Manager) ListTxs(ctx context.Context, params *ListTsxParams) (*ListTxsResult, error) {
	txs, err := man.repo.ListTxs(params.Cont)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	for i := range txs {
		if txs[i].Status == db.TxStatusPend && now.Sub(txs[i].Time) > time.Second*10 {
			err = man.loadTxStatus(ctx, &txs[i])
			if err == nil && txs[i].Status != db.TxStatusPend {
				txs[i].Time = time.Now()
				err = man.repo.UpdateTx(&txs[i])
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return &ListTxsResult{
		List: txs,
	}, nil
}

type SubmitTxParams struct {
	Server string
	Tx     *types.Transaction
}

func (man *Manager) SubmitTx(ctx context.Context, params *SubmitTxParams) error {
	client := proxy.NewClient(man.httpClient, params.Server, man.account)
	return client.SubmitTx(ctx, params.Tx)
}

func (man *Manager) loadTxStatus(ctx context.Context, obj *db.Tx) error {

	client := proxy.NewClient(man.httpClient, obj.Server, man.account)

	tx, pend, err := client.GetReceipt(ctx, common.HexToHash(obj.Hash))
	if err != nil {
		return err
	}

	obj.Time = time.Now()

	if tx == nil {
		obj.Status = db.TxStatusFail
	}
	if pend {
		obj.Status = db.TxStatusPend
	} else {
		obj.Status = db.TxStatusDone
	}
	return nil
}

func (man *Manager) loadServerInfo(ctx context.Context, server *db.Server) (err error) {

	client := proxy.NewClient(man.httpClient, server.URL, man.account)

	info, err := client.GetInfo(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	subs, err := client.GetSubs(ctx)
	if err != nil {
		return err
	}

	latency := time.Since(now)

	if info.Chain != man.chain {
		return errors.New("unknown server chain")
	}

	server.Version = info.Version
	server.Uptime = info.Uptime
	server.Conns = info.Conns
	server.Subs = info.Subs
	server.Height = info.Height
	server.Account = info.Account[:]
	server.GasPrice = info.GasPrice
	server.Balance = info.Balance
	server.RateFeed = info.RateFeed[:]
	server.RateSub = info.Rate.Sub.String()
	server.RateDiv = info.Rate.Div.String()
	server.RateMul = info.Rate.Mul.String()
	server.RateAdd = info.Rate.Add.String()
	server.Latency = latency.Milliseconds()
	server.Time = time.Now()

	server.MySubs = lo.Map(subs, func(sub proxy.Sub, _ int) db.Sub {
		return db.Sub{
			Tx:     sub.Tx.Hex(),
			Server: server.URL,
			Expire: sub.Expire,
			Remain: sub.Remain,
		}
	})

	return nil
}
