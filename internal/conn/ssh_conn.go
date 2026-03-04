package conn

import (
	"bufio"
	"context"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

type ShellConn struct {
	client  *ssh.Client
	session *ssh.Session

	stdin io.WriteCloser

	stdout, stderr *bufio.Reader
}

func NewShellConn(ctx context.Context, address, user, password string) (*ShellConn, error) {
	c, err := ssh.Dial("tcp", address, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}

	s, err := c.NewSession()
	if err != nil {
		return nil, err
	}

	stdin, err := s.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := s.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := s.Start("/bin/sh"); err != nil {
		return nil, err
	}

	stdoutBuf := bufio.NewReader(stdout)
	stderrBuf := bufio.NewReader(stderr)

	return &ShellConn{client: c, session: s, stdin: stdin, stdout: stdoutBuf, stderr: stderrBuf}, nil
}

func (s *ShellConn) Readout() (line string, err error) {
	if !s.ok() {
		return "", &connInvalidErr{}
	}
	line, err = s.stdout.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

func (s *ShellConn) Write(p []byte) (n int, err error) {
	if !s.ok() {
		return 0, &connInvalidErr{}
	}
	return s.stdin.Write(p)
}

func (s *ShellConn) Close() error {
	if !s.ok() {
		return &connInvalidErr{}
	}

	return s.session.Close()
}

func (s *ShellConn) ok() bool {
	return s != nil && s.session != nil && s.stdin != nil && s.stdout != nil && s.stderr != nil
}
