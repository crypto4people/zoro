package server

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type reporting sql.Stmt

func (server *Server) prepareReport(ctx context.Context, account common.Address) (*reporting, error) {
	query := "update subs set remain = remain - ?1 where id = (select id from subs where addr = '" + hex.EncodeToString(account[:]) + "' and remain > 0 and expire > unixepoch() order by expire asc limit 1)"
	prepared, err := server.db.conn.PrepareContext(ctx, query)
	return (*reporting)(prepared), err
}

func (rep *reporting) Close() error {
	return (*sql.Stmt)(rep).Close()
}

func (rep *reporting) Report(ctx context.Context, timeout time.Duration, dur time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := (*sql.Stmt)(rep).ExecContext(ctx, dur.Milliseconds()/1000)
	if err != nil {
		return err
	}
	affectd, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectd <= 0 {
		return errors.New("no active subscription")
	}
	return nil
}
