package service

import "context"

type Keyring struct {
}

func (kr *Keyring) Unlock(ctx context.Context, password string, fn func(key []byte) error) error {
}

func (ky *Keyring) Lock(ctx context.Context, key []byte, password string) error {
}
