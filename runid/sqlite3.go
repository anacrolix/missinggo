package runid

import (
	"context"
	"database/sql"

	"github.com/anacrolix/missinggo/expect"
)

type T int64

func New(db *sql.DB) (ret *T) {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	expect.Nil(err)
	defer func() {
		expect.Nil(conn.Close())
	}()
	_, err = conn.ExecContext(ctx, `CREATE TABLE if not exists runs (started datetime default (datetime('now')))`)
	expect.Nil(err)
	res, err := conn.ExecContext(ctx, "insert into runs default values")
	expect.Nil(err)
	expect.OneRowAffected(res)
	expect.Nil(conn.QueryRowContext(ctx, "select last_insert_rowid()").Scan(&ret))
	return
}
