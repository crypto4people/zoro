package server

import (
	"context"
)

func (server *Server) migrate(ctx context.Context) error {
	_, err := server.db.conn.ExecContext(ctx, `

		create table if not exists heights (
			chain text not null unique check (chain <> ''),
			height bigint not null check (height > 0)
		);

		create table if not exists subs (
			id integer primary key autoincrement,
			tx text not null unique check (tx <> ''),
			addr text check (tx <> ''),
			expire integer check (expire > 0),
			remain integer
		);

		create index if not exists idx_subs_actives on subs(addr, remain, expire) where addr is not null and remain is not null and remain > 0 and expire is not null;
		
	`)

	return err
}

func (server *Server) vaccum(ctx context.Context) error {
	_, err := server.db.conn.ExecContext(ctx, `
		
	VACUUM;
	
	`)
	return err
}
