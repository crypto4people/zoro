package service

import (
	"context"
	"crypto/ecdsa"
)

type Config struct {
	Key    *ecdsa.PrivateKey
	Server string
}

type Service struct {
}

func (serv *Service) Status(ctx context.Context) error {
	return nil
}

func (serv *Service) Start(ctx context.Context, config Config) error {
	return nil
}
