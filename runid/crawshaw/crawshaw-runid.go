package crawshaw_runid

import (
	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/anacrolix/missinggo/expect"
	"github.com/anacrolix/missinggo/v2/runid"
)

func New(db *sqlite.Conn) *runid.T {
	err := sqlitex.ExecScript(db, `
CREATE TABLE if not exists runs (started datetime default (datetime('now')));
insert into runs default values;
`)
	expect.Nil(err)
	ret := runid.T(db.LastInsertRowID())
	return &ret
}
