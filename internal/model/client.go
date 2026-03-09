package model

import (
	"context"
	"time"
)

// LogClient 日志客户端
type LogClient interface {
	Execute(ctx context.Context, pattern string, from, to time.Time) ResultStream
	SendMessage(ctx context.Context, message FrontendMessage) ResultStream
	ReceiveMessage(ctx context.Context) (BackendMessage, error)
	Close() error
	IsClosed() bool
	IsBusy() bool
	IsIdle() bool
}
