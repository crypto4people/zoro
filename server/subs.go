package server

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Sub struct {
	Expire int
	Total  int
	Remain int
}

func (server *Server) ListSubs(ctx context.Context, addr common.Address) ([]Sub, error) {
	if addr == (common.Address{}) {
		return []Sub{}, nil
	}

	rows, err := server.db.conn.QueryContext(ctx,
		"select expire, total, remain  from subs where addr = ?1 and remain > 0 and expire > unixepoch()",
		hex.EncodeToString(addr[:]),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Sub, 0, 4)

	var values [3]int

	for rows.Next() {
		err = rows.Scan(&values[0], &values[1], &values[2])
		if err != nil {
			return nil, err
		}
		result = append(result, Sub{
			Expire: values[0],
			Total:  values[1],
			Remain: values[2],
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

func (server *Server) createSub(ctx context.Context, tx common.Hash, addr common.Address, total int, ttl time.Duration) (bool, error) {

	result, err := server.db.conn.ExecContext(ctx,
		"insert into subs (tx, addr, total, remain, expire, chain) values (?1, ?2, ?3, ?3, unixepoch() + ?4, ?5) on conflict (tx) do nothing",
		hex.EncodeToString(tx[:]),
		hex.EncodeToString(addr[:]),
		total,
		ttl.Milliseconds()/1000,
		server.eth.chain.Text(10),
	)

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()

	return affected > 0, err
}
