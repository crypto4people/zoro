package service

import (
	"context"
	"errors"
	"time"
	"zoro/service/db"
	"zoro/service/db/model"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
)

type Storage struct {
	db *db.Database
}

var ErrNotFound = errors.New("not found")

func NewStorage(db *db.Database) *Storage {
	return &Storage{
		db: db,
	}
}

type AddServersParams struct {
	Server string
}

func (st *Storage) AddServers(ctx context.Context, params *AddServersParams) error {
	_, err := st.db.ServerBox.Put(&model.Server{
		URL:       params.Server,
		CreatedAt: time.Now(),
	})
	return err
}

type DeleteServerParams struct {
	Server string
}

func (st *Storage) DeleteServers(ctx context.Context, params *DeleteServerParams) error {
	return st.db.ServerBox.Remove(&model.Server{
		URL: params.Server,
	})
}

type UpsertServerInfoParams struct {
	Info *model.ServerInfo
}

func (st *Storage) UpsertServerInfo(ctx context.Context, params *UpsertServerInfoParams) error {
	params.Info.CreatedAt = time.Now()
	_, err := st.db.ServerInfoBox.Put(params.Info)
	return err
}

type ServerInfoParams struct {
	Server string
}

type ServerInfoResult struct {
	Info *model.ServerInfo
}

func (st *Storage) ServerInfo(ctx context.Context, params *ServerInfoParams) (*ServerInfoResult, error) {
	servers, err := st.db.ServerInfoBox.Query(
		model.ServerInfo_.URL.Equals(params.Server, false),
	).Find()
	if err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, ErrNotFound
	}

	return &ServerInfoResult{
		Info: servers[0],
	}, nil
}

type UpsertServerRateParams struct {
	Rate *model.ServerRate
}

func (st *Storage) UpsertServerRate(ctx context.Context, params *UpsertServerRateParams) error {
	params.Rate.CreatedAt = time.Now()
	_, err := st.db.ServerRateBox.Put(params.Rate)
	return err
}

type ServerRateParams struct {
	Server string
}

type ServerRateResult struct {
	Rate *model.ServerRate
}

func (st *Storage) ServerRate(ctx context.Context, params *ServerRateParams) (*ServerRateResult, error) {
	rates, err := st.db.ServerRateBox.Query(
		model.ServerRate_.URL.Equals(params.Server, false),
	).Find()
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, ErrNotFound
	}

	return &ServerRateResult{
		Rate: rates[0],
	}, nil
}

type UpsertServerSubsParams struct {
	Sub *model.ServerSubs
}

func (st *Storage) UpsertServerSubs(ctx context.Context, params *UpsertServerSubsParams) error {
	params.Sub.CreatedAt = time.Now()
	_, err := st.db.ServerSubsBox.Put(params.Sub)
	return err
}

type ServerSubsParams struct {
	Server  string
	Address common.Address
}

type ServerSubsResult struct {
	Subs []*model.ServerSubs
}

func (st *Storage) ServerSubs(ctx context.Context, params *ServerSubsParams) (*ServerSubsResult, error) {
	subs, err := st.db.ServerSubsBox.Query(
		model.ServerSubs_.URL.Equals(params.Server, false),
		model.ServerSubs_.Addr.Equals(params.Address.Bytes()),
	).Find()
	if err != nil {
		return nil, err
	}

	return &ServerSubsResult{
		Subs: subs,
	}, nil
}

type ListServersResult struct {
	Servers []string
}

func (st *Storage) ListServers(ctx context.Context) (*ListServersResult, error) {
	servers, err := st.db.ServerBox.GetAll()
	if err != nil || len(servers) == 0 {
		return nil, err
	}

	return &ListServersResult{
		Servers: lo.Map(servers, func(server *model.Server, _ int) string {
			return server.URL
		})}, nil
}

type AddTxParams struct {
	Server string
	Tx     *types.Transaction
}

func (st *Storage) AddTx(ctx context.Context, params *AddTxParams) error {
	txBytes, err := params.Tx.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = st.db.TxBox.Put(&model.Tx{
		Hash:      params.Tx.Hash().Bytes(),
		Server:    params.Server,
		Tx:        txBytes,
		Status:    "pending",
		CreatedAt: time.Now(),
	})
	return err
}

type UpdateTxStatusParams struct {
	TxId   uint64
	Status string
}

func (st *Storage) UpdateTxStatus(ctx context.Context, params *UpdateTxStatusParams) error {
	tx, err := st.db.TxBox.Get(params.TxId)
	if err != nil {
		return err
	}

	tx.Status = params.Status

	_, err = st.db.TxBox.Put(tx)
	return err
}

type ListTxsParams struct {
	Continuation uint64
}

type ListTxsResult struct {
	Txs          []*model.Tx
	Continuation uint64
}

func (st *Storage) ListTxs(ctx context.Context, params *ListTxsParams) (*ListTxsResult, error) {
	txs, err := st.db.TxBox.Query(
		model.Tx_.Id.GreaterThan(params.Continuation),
	).Limit(30).Find()
	if err != nil {
		return nil, err
	}

	return &ListTxsResult{
		Txs:          txs,
		Continuation: txs[len(txs)-1].Id,
	}, nil
}

// todo
func (st *Storage) RemoveOutDatedTxs(ctx context.Context, max time.Time) error
