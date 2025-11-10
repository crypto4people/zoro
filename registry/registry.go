package registry

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type Registry interface {
	List(ctx context.Context, requester common.Address) ([]string, error)
	Vote(ctx context.Context, voter common.Address, server string, vote bool) error
}
