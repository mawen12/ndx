package model

import (
	"fmt"
	"strings"
	"time"
)

type LogLine interface {
	Time() time.Time
	LogFile() string
	LogNumber() int
	Message() string
	Contexts() map[string]string
	Level() string
	OriginalLine() string
}

type QueryResult interface {
	Lines() []LogLine
	Statistics() map[int64]int
	Err() error
	Duration() time.Duration
}

type ResultStream interface {
	Read() QueryResult
}

type Notice struct {
	Conn     string
	Message  string
	Finished bool
	Success  bool
}

type KV struct {
	Key   string
	Value string
}

type ConnectMessages struct {
	Connected  []KV
	Connecting *ConnectMessage
}

func (cms ConnectMessages) String() string {
	var sb strings.Builder

	for _, kv := range cms.Connected {
		if kv.Value == "" {
			sb.WriteString(fmt.Sprintf(`%s [green::b]OK[-::-]`, kv.Key))
		} else {
			sb.WriteString(fmt.Sprintf(`%s [red::b]Failed[-::-]`, kv.Key))
			sb.WriteByte('\n')
			sb.WriteByte('\n')
			sb.WriteString(strings.Repeat(">", len(kv.Key)/2))
			sb.WriteString("[red::b]Errors[-::-]")
			sb.WriteString(strings.Repeat("<", len(kv.Key)/2))
			sb.WriteByte('\n')
			sb.WriteString(kv.Value)
		}

		sb.WriteByte('\n')
	}

	if cms.Connecting != nil {
		sb.WriteString(cms.Connecting.String())
		sb.WriteByte('\n')
	}

	return sb.String()
}

type ConnectMessage struct {
	Connection string
	Messages   []string
}

func (cm *ConnectMessage) String() string {
	var sb strings.Builder

	sb.WriteString(cm.Connection)

	for _, m := range cm.Messages {
		sb.WriteByte('\n')
		sb.WriteString(m)
	}

	return sb.String()
}
