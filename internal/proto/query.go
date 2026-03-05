package proto

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type Query struct {
	PathPrefix string
	ID         string
	LogFile    string
	From       string
	To         string
	Pattern    string
}

func (src *Query) Encode(dst []byte) ([]byte, error) {
	t := template.Must(template.New("bootstrap").Parse(querySh))
	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]any{
		"AgentPath":   fmt.Sprintf("%s/%s", src.PathPrefix, AgentShName),
		"IndexFile":   IndexFileName,
		"MaxNumLines": MaxNumLines,
		"LogFile":     src.LogFile,
		"FromExists":  src.From != "",
		"From":        src.From,
		"ToExists":    src.To != "",
		"To":          src.To,
		"Pattern":     shellQuote(src.Pattern),
		"ID":          src.ID,
	})
	if err != nil {
		return nil, err
	}

	bs := buf.Bytes()
	if bs[len(bs)-1] != '\n' {
		bs = append(bs, '\n')
	}

	dst = append(dst, bs...)
	return dst, nil
}

func shellQuote(s string) string {
	return fmt.Sprintf("'%s'", strings.Replace(s, "'", "'\"'\"'", -1))
}
