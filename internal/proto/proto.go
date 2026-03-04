package proto

import (
	_ "embed"
	"io"

	"github.com/mawen12/ndx/internal/model"
)

const (
	AgentShName       = "agent.sh"
	AgentLibShName    = "agent_lib.sh"
	AgentIndexShName  = "agent_index.sh"
	AgentSearchShName = "agent_search.sh"
	IndexFileName     = "index.log"

	MaxNumLines = 300
)

var (
	//go:embed scripts/startup.sh.tmpl
	startSh string
	//go:embed scripts/query.sh.tmpl
	querySh string
	//go:embed scripts/agent.sh
	agentSh string
	//go:embed scripts/agent_lib.sh
	libSh string
	//go:embed scripts/agent_index.sh
	indexSh string
	//go:embed scripts/agent_search.sh
	searchSh string
)

type Frontend struct {
	sr model.StringReader
	w  io.Writer

	startupMessage  StartupMessage
	dataLine        DataLine
	errorResponse   ErrorResponse
	stat            Stat
	noticeResponse  NoticeResponse
	parameterStatus ParameterStatus
	readyForQuery   ReadyForQuery
}

func NewFrontend(sr model.StringReader, w io.Writer) *Frontend {
	return &Frontend{sr: sr, w: w}
}

func (f *Frontend) Send(msg model.FrontendMessage) error {
	buf, err := msg.Encode(nil)
	if err != nil {
		return err
	}

	_, err = f.w.Write(buf)
	return err
}

func (f *Frontend) Receive() (model.BackendMessage, error) {
	line, err := f.sr.Next()

	var msg model.BackendMessage
	switch line[0] {
	case 'D':
		msg = &f.dataLine
	case 'E':
		msg = &f.errorResponse
	case 'T':
		msg = &f.stat
	case 'N':
		msg = &f.noticeResponse
	case 'S':
		msg = &f.parameterStatus
	case 'Z':
		msg = &f.readyForQuery
	}

	err = msg.Decode([]byte(line[1:]))

	return msg, err
}
