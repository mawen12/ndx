package pool

import (
	"context"
	"time"

	"github.com/mawen12/ndx/internal/conn"
)

type Conn struct {
	connString string
	*conn.NdxConn
	err error
}

func NewConn(ctx context.Context, connString string) *Conn {
	c, err := conn.Connect(ctx, connString)

	return &Conn{
		connString: connString,
		NdxConn:    c,
		err:        err,
	}
}

func (c *Conn) Reconnect(ctx context.Context) {
	if c.IsClosed() {
		con, err := conn.Connect(ctx, c.connString)

		c.NdxConn = con
		c.err = err
	}
}

func (c *Conn) Exec(ctx context.Context, pattern string, timeRange time.Time) *conn.Result {
	if c.IsClosed() {
		return &conn.Result{Err: c.err}
	}

	return c.NdxConn.Exec(ctx, pattern, timeRange).Read()
}
