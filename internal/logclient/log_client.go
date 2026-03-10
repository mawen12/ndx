package logclient

import (
	"context"
	"errors"
	"time"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/proto"
)

const (
	connStatusUninitialized = iota
	connStatusConnecting
	connStatusClosed
	connStatusIdle
	connStatusBusy
)

type LogClient struct {
	conn              model.Connection
	parameterStatuses map[string]string
	location          *time.Location
	frontend          *proto.Frontend

	config *Config

	status byte

	peekedMsg model.BackendMessage

	wbuf         []byte
	resultStream model.ResultStream

	cleanupDone chan struct{}
}

func Connect(ctx context.Context, qc *config.QueryConn) (*LogClient, error) {
	return ConnectConfig(ctx, NewConfig(qc))
}

func ConnectConfig(ctx context.Context, config *Config) (c *LogClient, err error) {
	if !config.CreatedByParseConfig {
		panic("config must be created by ParseConfig")
	}

	c, err = connect(ctx, config)
	if err != nil {
		return nil, &connectError{config: config, msg: "server error", err: err}
	}

	return c, nil
}

func connect(ctx context.Context, config *Config) (*LogClient, error) {
	c := new(LogClient)
	c.parameterStatuses = make(map[string]string)
	c.location = time.UTC
	c.config = config
	c.wbuf = make([]byte, 0, 8196)
	c.cleanupDone = make(chan struct{})

	conn, err := config.DialFunc(ctx, *config)
	if err != nil {
		return nil, &connectError{config: config, msg: "dial error", err: err}
	}

	c.conn = conn
	c.status = connStatusConnecting
	c.frontend = config.BuildFrontend(c.conn)

	startupMsg := &proto.StartupMessage{
		PathPrefix: c.config.PathPrefix,
		ID:         "1",
		LogFile:    config.LogFile,
	}

	if err := c.frontend.Send(startupMsg); err != nil {
		c.conn.Close()
		return nil, &connectError{config: config, msg: "failed to write startup message", err: err}
	}

	for {
		msg, err := c.receiveMessage()
		if err != nil {
			c.conn.Close()
			return nil, &connectError{config: config, msg: "failed to receive message", err: err}
		}

		switch msg := msg.(type) {
		case *proto.ReadyForQuery:
			c.status = connStatusIdle
			return c, nil
		case *proto.ErrorResponse:
			c.conn.Close()
			return nil, &connectError{config: config, msg: "handle startup message response failed", err: errors.New(msg.Message)}
		case *proto.ParameterStatus:
			if msg.Name == "tz" {
				location, err := time.LoadLocation(msg.Value)
				if err != nil {
					panic("not implemented")
				} else {
					c.location = location
				}
			}
		case *proto.NoticeResponse:
			// handle by receiveMessage
		default:
			c.conn.Close()
			return nil, &connectError{config: config, msg: "received unexpected message", err: err}
		}
	}
}

func (c *LogClient) Execute(ctx context.Context, pattern string, from, to time.Time) model.ResultStream {
	if err := c.lock(); err != nil {
		return &ResultStream{closed: true, err: err}
	}

	c.resultStream = &ResultStream{
		LogClient: c,
		ctx:       ctx,
	}

	queryMsg := &proto.Query{
		PathPrefix: c.config.PathPrefix,
		ID:         "1",
		LogFile:    c.config.LogFile,
		Pattern:    pattern,
	}

	if !from.IsZero() {
		queryMsg.From = from.In(c.location).Format("2006-01-02-15:04")
	}
	if !to.IsZero() {
		queryMsg.To = to.In(c.location).Format("2006-01-02-15:04")
	}

	if err := c.frontend.Send(queryMsg); err != nil {
		c.asyncClose()
		return &ResultStream{closed: true, err: err}
	}

	return c.resultStream
}

func (c *LogClient) SendMessage(ctx context.Context, message model.FrontendMessage) model.ResultStream {
	if err := c.lock(); err != nil {
		return &ResultStream{closed: true, err: err}
	}

	c.resultStream = &ResultStream{
		LogClient: c,
		ctx:       ctx,
	}

	if err := c.frontend.Send(message); err != nil {
		c.asyncClose()
		return &ResultStream{closed: true, err: err}
	}

	return c.resultStream
}

func (c *LogClient) ReceiveMessage(ctx context.Context) (model.BackendMessage, error) {
	if err := c.lock(); err != nil {
		return nil, err
	}
	defer c.unlock()

	if ctx != context.Background() {
		panic("not implemented")
	}

	msg, err := c.receiveMessage()
	if err != nil {
		panic("not implemented")
	}
	return msg, err
}

func (c *LogClient) receiveMessage() (model.BackendMessage, error) {
	msg, err := c.peekMessage()
	if err != nil {
		panic("not implemented")
	}

	c.peekedMsg = nil

	switch msg := msg.(type) {
	case *proto.ParameterStatus:
		c.parameterStatuses[msg.Name] = msg.Value
	case *proto.NoticeResponse:
		if c.config.OnNotice != nil {
			c.config.OnNotice(c, &Notice{Message: msg.Message})
		}
	case *proto.ErrorResponse:
		panic("not implemented")
	}

	return msg, err
}

func (c *LogClient) peekMessage() (model.BackendMessage, error) {
	if c.peekedMsg != nil {
		return c.peekedMsg, nil
	}

	msg, err := c.frontend.Receive()

	if err != nil {
		panic("not implemented")
	}

	c.peekedMsg = msg
	return msg, nil
}

func (c *LogClient) Close() error {
	if c.status == connStatusClosed {
		return nil
	}
	c.status = connStatusClosed

	defer close(c.cleanupDone)
	defer c.conn.Close()

	close := proto.Close{
		PathPrefix: c.config.PathPrefix,
	}
	buf, err := close.Encode(nil)
	if err == nil {
		c.conn.Write(buf)
	} else {
		return err
	}

	return c.conn.Close()
}

func (c *LogClient) asyncClose() {
	if c.status == connStatusClosed {
		return
	}
	c.status = connStatusClosed

	go func() {
		defer close(c.cleanupDone)
		defer c.conn.Close()

		close := proto.Close{
			PathPrefix: c.config.PathPrefix,
		}
		buf, err := close.Encode(nil)
		if err == nil {
			c.conn.Write(buf)
		}
	}()
}

func (c *LogClient) CleanupDone() chan struct{} {
	return c.cleanupDone
}

func (c *LogClient) IsClosed() bool {
	return c.status < connStatusIdle
}

func (c *LogClient) IsBusy() bool {
	return c.status == connStatusBusy
}

func (c *LogClient) IsIdle() bool {
	return c.status == connStatusIdle
}

func (c *LogClient) lock() error {
	switch c.status {
	case connStatusBusy:
		return &connLockError{status: "conn busy"}
	case connStatusClosed:
		return &connLockError{status: "conn closed"}
	case connStatusUninitialized:
		return &connLockError{status: "conn uninitialized"}
	default:
		c.status = connStatusBusy
		return nil
	}
}

func (c *LogClient) unlock() {
	switch c.status {
	case connStatusBusy:
		c.status = connStatusIdle
	case connStatusClosed:
	default:
		panic("BUG: cannot unlock unlocked connection")
	}
}
