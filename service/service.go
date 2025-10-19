package service

import (
	"context"
	"crypto/ecdsa"
)

type ProxyServiceConfig struct {
	Key    *ecdsa.PrivateKey
	Server string
}

func StartProxyService(ctx context.Context, config ProxyServiceConfig) error {
	return nil
}
