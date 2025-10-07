package server

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/core/types"
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
		return 0, errors.New("block not exists")
	}

	return server.processBlock(ctx, number)
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

func (server *Server) processBlock(ctx context.Context, number int64) (int, error) {

	block, err := server.eth.client.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return 0, err
	}

	rate, err := server.rate.feed(ctx, number)
	if err != nil {
		return 0, err
	}

	changes := 0
	for _, tx := range block.Transactions() {
		applied, err := server.processTx(ctx, &rate, tx)
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

	convertResult := rate.convert(value)

	created, err := server.createSub(ctx, tx.Hash(), sender, convertResult.Dur, convertResult.TTL)
	if err != nil {
		return false, err
	}

	if created {
		if convertResult.TTL == 0 {
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
				slog.Duration("dur", convertResult.Dur),
				slog.Duration("ttl", convertResult.TTL),
			)
		}
	} else {
		server.logger.Info("duplicate tx",
			slog.String("tx", tx.Hash().Hex()),
		)
	}

	return created, nil
}
