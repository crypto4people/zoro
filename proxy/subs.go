package proxy

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Sub struct {
	Tx     common.Hash
	Expire int64
	Remain int64
}

func (server *Server) ListSubs(ctx context.Context, addr common.Address) ([]Sub, error) {
	if addr == (common.Address{}) {
		return []Sub{}, nil
	}

	rows, err := server.db.conn.QueryContext(ctx,
		"select tx, expire, remain from subs where addr = ?1 and remain > 0 and expire > unixepoch() limit 20",
		hex.EncodeToString(addr[:]),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Sub, 0, 4)

	var tx string
	var values [2]int64

	for rows.Next() {
		err = rows.Scan(&tx, &values[0], &values[1])
		if err != nil {
			return nil, err
		}

		result = append(result, Sub{
			Tx:     common.HexToHash(tx),
			Expire: values[0],
			Remain: values[1],
		})
	}

	return result, nil
}

func (server *Server) countSubs(ctx context.Context) error {
	var count int
	err := server.db.conn.QueryRowContext(ctx, "select count(distinct addr) from subs where remain > 0 and expire > unixepoch()").Scan(&count)
	if err != nil {
		return err
	}

	server.db.subs.Store(int64(count))
	return nil
}

func (server *Server) createSub(ctx context.Context, tx common.Hash, addr common.Address, dur time.Duration, ttl time.Duration) (bool, error) {
	var err error
	var result sql.Result
	if ttl == 0 {
		result, err = server.db.conn.ExecContext(ctx,
			"insert into subs (tx) values (?1) on conflict (tx) do nothing",
			hex.EncodeToString(tx[:]),
		)
	} else {
		result, err = server.db.conn.ExecContext(ctx,
			"insert into subs (tx, addr, remain, expire) values (?1, ?2, ?3, unixepoch() + ?4) on conflict (tx) do nothing",
			hex.EncodeToString(tx[:]),
			hex.EncodeToString(addr[:]),
			dur.Milliseconds()/1000,
			ttl.Milliseconds()/1000,
		)
	}

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()

	return affected > 0, err
}
