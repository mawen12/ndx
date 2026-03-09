package pool

import (
	"context"
	"time"

	"github.com/mawen12/ndx/internal/logclient"
	"github.com/mawen12/ndx/internal/model"
)

type Conn struct {
	model.LogClient
	connString string
	err        error
}

func NewConn(ctx context.Context, connString string) *Conn {
	c, err := logclient.Connect(ctx, connString)

	return &Conn{
		connString: connString,
		LogClient:  c,
		err:        err,
	}
}

func (c *Conn) Reconnect(ctx context.Context) {
	if c.IsClosed() {
		con, err := logclient.Connect(ctx, c.connString)

		c.LogClient = con
		c.err = err
	}
}

func (c *Conn) Exec(ctx context.Context, pattern string, from, to time.Time) model.QueryResult {
	//if c.IsClosed() {
	//	return &logclient.QueryResult{err: c.err}
	//}

	return c.LogClient.Execute(ctx, pattern, from, to).Read()
}
