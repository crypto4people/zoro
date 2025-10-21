package service

import (
	"context"
	"zoro/service/db/model"

	"github.com/samber/lo"
)

type Storage struct {
	serverBox *model.ServerBox
}

func NewStorage(serverBox *model.ServerBox) *Storage {
	return &Storage{
		serverBox: serverBox,
	}
}

func (st *Storage) AddServers(ctx context.Context, server string) error {
	_, err := st.serverBox.Put(&model.Server{
		URL: server,
	})
	return err
}

func (st *Storage) DeleteServers(ctx context.Context, server string) error {
	return st.serverBox.Remove(&model.Server{
		URL: server,
	})
}

func (st *Storage) ListServers(ctx context.Context) []string {
	servers, err := st.serverBox.GetAll()
	if err != nil || len(servers) == 0 {
		return []string{}
	}

	return lo.Map(servers, func(server *model.Server, _ int) string {
		return server.URL
	})
}
