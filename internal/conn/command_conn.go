package conn

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
)

type CommandConn struct {
	cmd *exec.Cmd

	stdin          io.WriteCloser
	stdout, stderr io.ReadCloser

	stdoutBuf, stderrBuf *bufio.Reader
}

func NewCommandConn(ctx context.Context, command string, arg ...string) (*CommandConn, error) {
	cmd := exec.CommandContext(ctx, command, arg...)

	// TODO close when error occur
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	// TODO close when error occur
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// TODO close when error occur
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	stdoutBuf := bufio.NewReader(stdout)
	stderrBuf := bufio.NewReader(stderr)

	return &CommandConn{cmd: cmd, stdin: stdin, stdout: stdout, stderr: stderr, stdoutBuf: stdoutBuf, stderrBuf: stderrBuf}, nil
}

func (c *CommandConn) Readout() (line string, err error) {
	if !c.ok() {
		return "", &connInvalidErr{}
	}

	line, err = c.stdoutBuf.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

func (c *CommandConn) Write(b []byte) (n int, err error) {
	if !c.ok() {
		return 0, &connInvalidErr{}
	}
	return c.stdin.Write(b)
}

func (c *CommandConn) Close() error {
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	_ = c.stderr.Close()
	return c.cmd.Wait()
}

func (c *CommandConn) ok() bool {
	return c != nil && c.cmd != nil && c.stdin != nil && c.stdout != nil && c.stderr != nil
}
