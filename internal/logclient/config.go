package logclient

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/conn"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/proto"
)

type DialFunc func(ctx context.Context, config Config) (model.Connection, error)

type BuildFrontendFunc func(conn model.Connection) *proto.Frontend

type Notice struct {
	Message string
}

type NoticeHandler func(*LogClient, *Notice)

type Config struct {
	*config.QueryConn

	Command       string
	DialFunc      DialFunc
	BuildFrontend BuildFrontendFunc
	RuntimeParams map[string]string

	OnNotice             NoticeHandler
	PathPrefix           string
	CreatedByParseConfig bool
}

func NewConfig(qc *config.QueryConn) *Config {
	return &Config{
		QueryConn:            qc,
		Command:              "/bin/sh",
		RuntimeParams:        make(map[string]string),
		DialFunc:             makeDialFunc(qc.Scheme, qc.Host),
		BuildFrontend:        makeDefaultBuildFrontendFunc(),
		PathPrefix:           fmt.Sprintf("/tmp/ndx_%s", uuid.New().String()),
		CreatedByParseConfig: true,
	}
}

func makeDefaultBuildFrontendFunc() BuildFrontendFunc {
	return func(conn model.Connection) *proto.Frontend {
		return proto.NewFrontend(conn)
	}
}

func makeDialFunc(scheme string, host string) DialFunc {
	switch scheme {
	case "cmd":
		return func(ctx context.Context, config Config) (model.Connection, error) {
			return NewCmdConnConfig(ctx, config)
		}
	case "ssh":
		return func(ctx context.Context, config Config) (model.Connection, error) {
			return NewShellConnConfig(ctx, config)
		}
	default:
		panic("BUG: unsupported scheme for makeDialFunc")
	}
}

func NewCmdConnConfig(ctx context.Context, config Config) (*conn.CommandConn, error) {
	return conn.NewCommandConn(ctx, config.Command)
}

func NewShellConnConfig(ctx context.Context, config Config) (*conn.SSHConn, error) {
	return conn.NewSSHConn(ctx, fmt.Sprintf("%s:%d", config.Host, int(config.Port)), config.User, config.Password)
}
