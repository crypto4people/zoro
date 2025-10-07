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
			chain text not null check (chain <> ''),
			addr text not null check (addr <> ''),
			time integer not null default (unixepoch()) check (time > 0),
			total integer not null check (total > 0),
			expire integer not null check (expire > 0),
			remain integer not null
		);

		create index if not exists idx_subs_actives on subs(addr, remain, expire) where remain > 0;
		
	`)

	return err
}

func (server *Server) vaccum(ctx context.Context) error {
	_, err := server.db.conn.ExecContext(ctx, `
		
	VACUUM;
	
	`)
	return err
}
