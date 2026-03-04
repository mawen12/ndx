package proto

import (
	"bytes"
	"text/template"
)

type StartupMessage struct {
	PathPrefix string
	ID         string
	LogFile    string
}

func (src *StartupMessage) Encode(dst []byte) ([]byte, error) {
	t := template.Must(template.New("bootstrap").Parse(startSh))
	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]any{
		"PrefixPath":    src.PathPrefix,
		"AgentPath":     AgentShName,
		"AgentContent":  agentSh,
		"LibPath":       AgentLibShName,
		"LibContent":    libSh,
		"IndexPath":     AgentIndexShName,
		"IndexContent":  indexSh,
		"SearchPath":    AgentSearchShName,
		"SearchContent": searchSh,
		"IndexFile":     IndexFileName,
		"LogFile":       src.LogFile,
		"Params":        "",
		"ID":            src.ID,
	})
	if err != nil {
		return nil, err
	}

	bs := buf.Bytes()
	if bs[len(bs)-1] != '\n' {
		bs = append(bs, '\n')
	}

	dst = append(dst, buf.Bytes()...)
	return dst, nil
}
