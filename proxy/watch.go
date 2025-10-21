package proxy

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (server *Server) Watch(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

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
			duration := server.ethClient.BlockTime()
			server.logger.Debug("watcher sleep", slog.Duration("duration", duration))

			timer.Reset(duration)

			if err := server.vacuum(ctx); err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				continue
			}
		}

		for number := localHeight + 1; number <= networkHeight; number++ {
			changes, err := server.processBlock(ctx, number)
			if err != nil {
				return err
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
		return 0, errors.New("block not finalized")
	}

	return server.processBlock(ctx, number)
}

func (server *Server) setLocalHeight(ctx context.Context, height int64) error {
	_, err := server.db.conn.ExecContext(ctx, "insert into heights(chain,height) values(?1,?2) on conflict(chain) do update set height = ?2 where height < ?2", server.ethClient.Chain(), height)
	if err == nil {
		server.db.height.Store(height)
	}

	return err
}

func (server *Server) getLocalHeight(ctx context.Context) (int64, error) {
	var height int64

	err := server.db.conn.QueryRowContext(ctx, "select height from heights where chain = ?1", server.ethClient.Chain()).Scan(&height)
	if err == nil {
		server.db.height.Store(height)
	}

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return height, err
}

func (server *Server) getNetworkHeight(ctx context.Context) (int64, error) {
	return server.ethClient.FinalHeight(ctx)
}

func (server *Server) processBlock(ctx context.Context, number int64) (int, error) {

	rate, err := server.ethClient.RateAt(ctx, number)
	if err != nil {
		return 0, err
	}

	changes := 0

	txs, err := server.ethClient.Receives(ctx, number)
	if err != nil {
		return 0, err
	}

	for sender, tx := range txs {
		applied, err := server.processTx(ctx, rate, sender, tx)
		if err != nil {
			return changes, err
		}
		if applied {
			changes += 1
		}
	}

	server.logger.Debug("block processes",
		slog.Int64("number", number),
	)

	return changes, nil
}

func (server *Server) processTx(ctx context.Context, rate *rate.Rate, sender common.Address, tx *types.Transaction) (bool, error) {

	value := tx.Value()

	dur, ttl := rate.Apply(value)

	created, err := server.createSub(ctx, tx.Hash(), sender, dur, ttl)
	if err != nil {
		return false, err
	}

	if created {
		if ttl == 0 {
			server.logger.Info("sub rejected",
				slog.String("tx", tx.Hash().Hex()),
				slog.String("addr", sender.Hex()),
				slog.String("value", value.Text(10)),
			)
		} else {
			server.logger.Info("sub created",
				slog.String("tx", tx.Hash().Hex()),
				slog.String("addr", sender.Hex()),
				slog.String("value", value.Text(10)),
				slog.Duration("dur", dur),
				slog.Duration("ttl", ttl),
			)
		}
	} else {
		server.logger.Info("duplicate tx",
			slog.String("tx", tx.Hash().Hex()),
		)
	}

	return created, nil
}
