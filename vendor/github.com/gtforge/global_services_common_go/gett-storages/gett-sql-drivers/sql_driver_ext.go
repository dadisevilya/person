package gettSqlDrivers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
	"sync/atomic"
)

type connExt struct {
	driver.Conn
	driver.ConnBeginTx
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext

	driver *driverExt
}

func (conn connExt) Begin() (driver.Tx, error) {
	var ctx = context.Background()
	return conn.ConnBeginTx.BeginTx(ctx, conn.driver.txDefaults)
}

func (conn connExt) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return conn.ConnBeginTx.BeginTx(ctx, conn.driver.txDefaults)
}

type driverExt struct {
	driverName string
	txDefaults driver.TxOptions

	init     sync.Once
	inner    *sql.DB
	innerErr error
}

func (d *driverExt) Open(dataSourceName string) (conn driver.Conn, connErr error) {
	d.init.Do(func() {
		d.inner, d.innerErr = sql.Open("postgres", dataSourceName)
	})
	if d.innerErr != nil {
		return nil, d.innerErr
	}
	conn, connErr = d.inner.Driver().Open(dataSourceName)
	conn = connExt{
		Conn:           conn,
		ConnBeginTx:    conn.(driver.ConnBeginTx),
		Execer:         conn.(driver.Execer),
		ExecerContext:  conn.(driver.ExecerContext),
		Queryer:        conn.(driver.Queryer),
		QueryerContext: conn.(driver.QueryerContext),
		driver:         d,
	}
	return
}

func NewDriver(upstream string, txDefaults *driver.TxOptions) string {
	var key = fmt.Sprintf("%s_%d", upstream, atomic.AddUint64(&regCounter, 1))
	var d = &driverExt{
		driverName: upstream,
		txDefaults: driver.TxOptions{},
	}
	if txDefaults != nil {
		d.txDefaults = *txDefaults
	}
	sql.Register(key, d)
	return key
}

var regCounter uint64
