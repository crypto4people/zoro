package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/gofrs/flock"
)

func readSharedJSON[T any](ctx context.Context, path string) (*T, error) {
	lock := flock.New(path)
	defer lock.Close()
	locked, err := lock.TryRLockContext(ctx, time.Second)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("failed to lock file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v T
	err = json.Unmarshal(data, &v)
	return &v, err
}

func writeExclusiveJSON(ctx context.Context, path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	lock := flock.New(path)
	defer lock.Close()
	locked, err := lock.TryLockContext(ctx, time.Second)
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("failed to lock file")
	}
	return os.WriteFile(path, data, 0644)
}
