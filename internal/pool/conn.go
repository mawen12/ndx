package pool

import (
	"context"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/logclient"
	"github.com/mawen12/ndx/internal/model"
)

type Conn struct {
	*logclient.LogClient
	Conn *config.QueryConn
}

func NewConn(conn *config.QueryConn) *Conn {
	return &Conn{
		Conn: conn,
	}
}

func (c *Conn) Connect(ctx context.Context) error {
	client, err := logclient.Connect(ctx, c.Conn)
	if err != nil {
		return err
	}

	c.LogClient = client
	return nil
}

func (c *Conn) Exec(ctx context.Context, queryContext model.QueryContext) model.QueryResult {
	return c.LogClient.Execute(ctx, queryContext).Read()
}
