package runid

import (
	"context"
	"database/sql"

	"github.com/anacrolix/missinggo/assert"
)

type T int64

func New(db *sql.DB) (ret *T) {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	assert.Nil(err)
	defer func() {
		assert.Nil(conn.Close())
	}()
	res, err := conn.ExecContext(ctx, "insert into runs default values")
	assert.Nil(err)
	assert.OneRowAffected(res)
	assert.Nil(conn.QueryRowContext(ctx, "select last_insert_rowid()").Scan(&ret))
	return
}
