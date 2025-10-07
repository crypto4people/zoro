package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/itsabgr/ge"
	"github.com/itsabgr/retry"
)

func (server *Server) Watch(ctx context.Context) error {
	for {

		networkHeight, err := server.getNetworkHeight(ctx)
		if err != nil {
			return err
		}
		localHeight, err := server.getLocalHeight(ctx)
		if err != nil {
			return err
		}
		if localHeight > networkHeight {
			return errors.New("old network height")
		}

		if localHeight == networkHeight {
			server.logger.Debug("watcher sleep", slog.Duration("duration", server.eth.sleep))
			if err = retry.Sleep(ctx, server.eth.sleep); err != nil {

				return err
			}
			continue
		}

		for number := localHeight + 1; number <= networkHeight; number++ {
			changes, err := server.tryProcessBlock(ctx, number)
			if err != nil {
				continue
			}
			if err = server.setLocalHeight(ctx, number); err != nil {
				return err
			}
			if changes > 0 {
				if err = server.countSubs(ctx); err != nil {
					return err
				}
			}
		}

	}
}

func (server *Server) ProcessBlock(ctx context.Context, number int64) (changes int, err error) {
	networkHeight, err := server.getNetworkHeight(ctx)
	if err != nil {
		return 0, err
	}
	if networkHeight < number {
		return 0, errors.New("block not exists")
	}

	return server.tryProcessBlock(ctx, number)
}

func (server *Server) setLocalHeight(ctx context.Context, height int64) error {
	_, err := server.db.conn.ExecContext(ctx, "insert into heights(chain,height) values(?1,?2) on conflict(chain) do update set height = ?2 where height < ?2", server.eth.chain.Text(10), height)
	if err == nil {
		server.db.height.Store(height)
	}
	return err
}

func (server *Server) getLocalHeight(ctx context.Context) (int64, error) {
	var height int64
	err := server.db.conn.QueryRowContext(ctx, "select height from heights where chain = ?1", server.eth.chain.Text(10)).Scan(&height)
	if err == sql.ErrNoRows {
		err = nil
		height = 0
	}
	if err == nil {
		server.db.height.Store(height)
	}

	return max(height, server.eth.start), err
}

func (server *Server) getNetworkHeight(ctx context.Context) (int64, error) {
	number, err := server.eth.client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	if number <= 0 || number > math.MaxInt64 {
		return 0, errors.New("invalid network height")
	}

	return max(int64(number)-int64(server.eth.offset), 1), nil
}

func (server *Server) tryProcessBlock(ctx context.Context, number int64) (int, error) {
	rate, err := server.rate.feed(ctx, number)
	if err != nil {
		return 0, err
	}
	block, err := server.eth.client.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return 0, err
	}
	changes := 0
	errors := make([]error, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		applied, err := server.processTx(ctx, &rate, tx)
		if err != nil {
			server.logger.Warn("failed to process tx",
				slog.Int64("block", number),
				slog.String("tx", tx.Hash().Hex()),
				slog.String("err", err.Error()),
			)
			errors = append(errors, err)
			continue
		}
		if applied {
			changes += 1
		}
	}
	server.logger.Debug("new block processes",
		slog.Int64("block", number),
	)
	return changes, ge.Join(errors...)
}

func (server *Server) processTx(ctx context.Context, rate *Rate, tx *types.Transaction) (bool, error) {
	to := tx.To()
	value := tx.Value()
	if to == nil || value.Sign() <= 0 || len(tx.Data()) > 0 || !slices.Contains(server.eth.addrs, *to) {
		return false, nil
	}

	sender, err := server.eth.signer.Sender(tx)
	if err != nil {
		return false, err
	}

	gas, ttl := rate.Convert(value)
	if gas <= 0 || ttl <= 0 {
		fmt.Println(value)
		return false, fmt.Errorf("failed to convert value %v", value)
	}

	created, err := server.createSub(ctx, tx.Hash(), sender, gas, ttl)
	if err != nil {
		return false, err
	}

	if created {
		server.logger.Info("new sub created",
			slog.String("tx", tx.Hash().Hex()),
			slog.String("addr", sender.Hex()),
			slog.String("value", value.Text(10)),
			slog.Int("gas", gas),
			slog.Duration("ttl", ttl),
		)
	} else {
		server.logger.Info("duplicate tx",
			slog.String("tx", tx.Hash().Hex()),
		)
	}

	return created, nil
}
