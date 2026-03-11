package pool

import (
	"context"
	"time"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/logclient"
	"github.com/mawen12/ndx/internal/model"
)

type Notice struct {
	Message string
	Err     error
}

type Conn struct {
	*logclient.LogClient
	Conn *config.QueryConn
}

func NewConn(conn *config.QueryConn) *Conn {
	return &Conn{
		Conn: conn,
	}
}

func (c *Conn) Connect(ctx context.Context, callback func(string, bool)) error {
	client, err := logclient.Connect(ctx, c.Conn, callback)
	if err != nil {
		return err
	}

	c.LogClient = client
	return nil
}

func (c *Conn) Exec(ctx context.Context, pattern string, from, to time.Time) model.QueryResult {
	return c.LogClient.Execute(ctx, pattern, from, to).Read()
}
