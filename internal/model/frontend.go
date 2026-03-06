package model

// Frontend 定义前端消息收发协议
type Frontend interface {
	// Send 发送前端消息
	Send(msg FrontendMessage) error

	// Receive 接收后端消息
	Receive() (BackendMessage, error)
}
