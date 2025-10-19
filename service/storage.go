package service

import (
	"context"
)

type Storage struct {
}

func (st *Storage) AddServers(ctx context.Context, server string) error {
	panic("unimplemented")
}

func (st *Storage) DeleteServers(ctx context.Context, server string) error {
	panic("unimplemented")
}

func (st *Storage) ListServers(ctx context.Context) []string {
	panic("unimplemented")
}
