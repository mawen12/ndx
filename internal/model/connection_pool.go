package model

import (
	"context"
)

// PoolStatistics 表示连接池的统计信息
type PoolStatistics interface {
	Idle() int
	Busy() int
	Closed() int
}

// PoolStatisticsListener 连接池统计信息管理器
type PoolStatisticsListener func(PoolStatistics)

// ConnectionPool 管理多个 LogClient 连接，负责多源并发查询和连接状态管理
type ConnectionPool interface {
	// Query 向所有可用客户端发送查询请求
	Query(ctx context.Context, qc *QueryContext) *ResultStream

	// AddStatisticsListener 添加统计监听器
	AddStatisticsListener(PoolStatisticsListener)

	// RemoveStatisticsListener 移除统计信息监听器
	RemoveStatisticsListener(PoolStatisticsListener)

	// Statistics 获取当前连接池统计信息
	Statistics() PoolStatistics

	// Close 关闭所有客户端连接
	Close()
}
