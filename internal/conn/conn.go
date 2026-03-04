package conn

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

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

const wbufLen = 1024

type DialFunc func(ctx context.Context, config Config) (model.Conn, error)

type BuildFrontendFunc func(conn model.Conn, w io.Writer) Frontend

type NoticeHandler func(*NdxConn, *Notice)

type Frontend interface {
	Send(msg model.FrontendMessage) error
	Receive() (model.BackendMessage, error)
}

type Notice struct {
	Message string
}

type NdxConn struct {
	conn              model.Conn
	parameterStatuses map[string]string
	location          *time.Location
	frontend          Frontend

	config *Config

	status byte

	bufferingReceive    bool
	bufferingReceiveMux sync.Mutex
	bufferingReceiveMsg model.BackendMessage
	bufferingReceiveErr error

	peekedMsg model.BackendMessage

	wbuf         []byte
	resultReader ResultReader

	cleanupDone chan struct{}
}

func Connect(ctx context.Context, connString string) (*NdxConn, error) {
	config, err := ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	return ConnectConfig(ctx, config)
}

func ConnectConfig(ctx context.Context, config *Config) (ndConn *NdxConn, err error) {
	if !config.createdByParseConfig {
		panic("config must be created by ParseConfig")
	}

	ndConn, err = connect(ctx, config)
	if err != nil {
		return nil, &connectError{config: config, msg: "server error", err: err}
	}

	return ndConn, nil
}

func connect(ctx context.Context, config *Config) (*NdxConn, error) {
	ndConn := new(NdxConn)
	ndConn.parameterStatuses = make(map[string]string)
	ndConn.location = time.UTC
	ndConn.config = config
	ndConn.wbuf = make([]byte, 0, wbufLen)
	ndConn.cleanupDone = make(chan struct{})

	conn, err := config.DialFunc(ctx, *config)
	if err != nil {
		return nil, &connectError{config: config, msg: "dial error", err: err}
	}

	ndConn.conn = conn
	ndConn.status = connStatusConnecting
	ndConn.frontend = config.BuildFrontend(ndConn.conn, ndConn.conn)

	startupMsg := &proto.StartupMessage{
		PathPrefix: ndConn.config.PathPrefix,
		ID:         "1",
		LogFile:    config.LogFile,
	}

	//buf, err := startupMsg.Encode(ndConn.wbuf)
	//if err != nil {
	//	return nil, &connectError{config: config, msg: "failed to write startup message", err: err}
	//}

	if err := ndConn.frontend.Send(startupMsg); err != nil {
		ndConn.conn.Close()
		return nil, &connectError{config: config, msg: "failed to write startup message", err: err}
	}

	for {
		msg, err := ndConn.receiveMessage()
		if err != nil {
			ndConn.conn.Close()
			return nil, &connectError{config: config, msg: "failed to receive message", err: err}
		}

		switch msg := msg.(type) {
		case *proto.ReadyForQuery:
			ndConn.status = connStatusIdle
			return ndConn, nil
		case *proto.ErrorResponse:
			ndConn.conn.Close()
			return nil, &connectError{config: config, msg: "handle startup message response failed", err: errors.New(msg.Message)}
		case *proto.ParameterStatus:
			if msg.Name == "tz" {
				location, err := time.LoadLocation(msg.Value)
				if err != nil {
					panic("not implemented")
				} else {
					ndConn.location = location
				}
			}
		case *proto.NoticeResponse:
			// handle by receiveMessage
		default:
			ndConn.conn.Close()
			return nil, &connectError{config: config, msg: "received unexpected message", err: err}
		}
	}
}

func (ndConn *NdxConn) Exec(ctx context.Context, pattern string, timeRange time.Time) *ResultReader {
	if err := ndConn.lock(); err != nil {
		return &ResultReader{closed: true, err: err}
	}

	ndConn.resultReader = ResultReader{
		ndConn: ndConn,
		ctx:    ctx,
	}
	result := &ndConn.resultReader

	queryMsg := &proto.Query{
		PathPrefix: ndConn.config.PathPrefix,
		ID:         "1",
		LogFile:    ndConn.config.LogFile,
		Pattern:    pattern,
	}

	if err := ndConn.frontend.Send(queryMsg); err != nil {
		ndConn.asyncClose()
		result.err = err
		return result
	}

	return result
}

func (ndConn *NdxConn) ReceiveMessage(ctx context.Context) (model.BackendMessage, error) {
	if err := ndConn.lock(); err != nil {
		return nil, err
	}
	defer ndConn.unlock()

	if ctx != context.Background() {
		panic("not implemented")
	}

	msg, err := ndConn.receiveMessage()
	if err != nil {
		panic("not implemented")
	}
	return msg, err
}

func (ndConn *NdxConn) receiveMessage() (model.BackendMessage, error) {
	msg, err := ndConn.peekMessage()
	if err != nil {
		panic("not implemented")
	}

	ndConn.peekedMsg = nil

	switch msg := msg.(type) {
	case *proto.ParameterStatus:
		ndConn.parameterStatuses[msg.Name] = msg.Value
	case *proto.NoticeResponse:
		if ndConn.config.OnNotice != nil {
			ndConn.config.OnNotice(ndConn, &Notice{Message: msg.Message})
		}
	case *proto.ErrorResponse:
		panic("not implemented")
	}

	return msg, err
}

func (ndConn *NdxConn) peekMessage() (model.BackendMessage, error) {
	if ndConn.peekedMsg != nil {
		return ndConn.peekedMsg, nil
	}

	var msg model.BackendMessage
	var err error

	if ndConn.bufferingReceive {
		ndConn.bufferingReceiveMux.Lock()
		msg = ndConn.bufferingReceiveMsg
		err = ndConn.bufferingReceiveErr
		ndConn.bufferingReceiveMux.Unlock()
		ndConn.bufferingReceive = false

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			msg, err = ndConn.frontend.Receive()
		}

	} else {
		msg, err = ndConn.frontend.Receive()
	}

	if err != nil {
		panic("not implemented")
	}

	ndConn.peekedMsg = msg
	return msg, nil
}

func (ndConn *NdxConn) Close() error {
	if ndConn.status == connStatusClosed {
		return nil
	}
	ndConn.status = connStatusClosed

	defer close(ndConn.cleanupDone)
	defer ndConn.conn.Close()

	close := proto.Close{
		PathPrefix: ndConn.config.PathPrefix,
	}
	buf, err := close.Encode(nil)
	if err == nil {
		ndConn.conn.Write(buf)
	} else {
		return err
	}

	return ndConn.conn.Close()
}

func (ndConn *NdxConn) asyncClose() {
	if ndConn.status == connStatusClosed {
		return
	}
	ndConn.status = connStatusClosed

	go func() {
		defer close(ndConn.cleanupDone)
		defer ndConn.conn.Close()

		close := proto.Close{
			PathPrefix: ndConn.config.PathPrefix,
		}
		buf, err := close.Encode(nil)
		if err == nil {
			ndConn.conn.Write(buf)
		}
	}()
}

func (ndConn *NdxConn) CleanupDone() chan struct{} {
	return ndConn.cleanupDone
}

func (ndConn *NdxConn) IsClosed() bool {
	return ndConn.status < connStatusIdle
}

func (ndConn *NdxConn) IsBusy() bool {
	return ndConn.status == connStatusBusy
}

func (ndConn *NdxConn) lock() error {
	switch ndConn.status {
	case connStatusBusy:
		return &connLockError{status: "conn busy"}
	case connStatusClosed:
		return &connLockError{status: "conn closed"}
	case connStatusUninitialized:
		return &connLockError{status: "conn uninitialized"}
	default:
		ndConn.status = connStatusBusy
		return nil
	}
}

func (ndConn *NdxConn) unlock() {
	switch ndConn.status {
	case connStatusBusy:
		ndConn.status = connStatusIdle
	case connStatusClosed:
	default:
		panic("BUG: cannot unlock unlocked connection")
	}
}
