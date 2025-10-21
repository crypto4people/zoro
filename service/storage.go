package service

import (
	"context"
	"zoro/service/db"
	"zoro/service/db/model"

	"github.com/samber/lo"
)

type Storage struct {
	db *db.Database
}

func NewStorage(db *db.Database) *Storage {
	return &Storage{
		db: db,
	}
}

func (st *Storage) AddServers(ctx context.Context, server string) error {
	_, err := st.db.ServerBox.Put(&model.Server{
		URL: server,
	})
	return err
}

func (st *Storage) DeleteServers(ctx context.Context, server string) error {
	return st.db.ServerBox.Remove(&model.Server{
		URL: server,
	})
}

func (st *Storage) ListServers(ctx context.Context) []string {
	servers, err := st.db.ServerBox.GetAll()
	if err != nil || len(servers) == 0 {
		return []string{}
	}

	return lo.Map(servers, func(server *model.Server, _ int) string {
		return server.URL
	})
}
