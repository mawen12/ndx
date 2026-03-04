package conn

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
)

type CmdConn struct {
	cmd *exec.Cmd

	stdin io.WriteCloser

	stdout, stderr *bufio.Reader
}

func NewCmdConn(ctx context.Context, command string, arg ...string) (*CmdConn, error) {
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

	return &CmdConn{cmd: cmd, stdin: stdin, stdout: stdoutBuf, stderr: stderrBuf}, nil
}

func (c *CmdConn) Readout() (line string, err error) {
	if !c.ok() {
		return "", &connInvalidErr{}
	}

	line, err = c.stdout.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

func (c *CmdConn) Write(b []byte) (n int, err error) {
	if !c.ok() {
		return 0, &connInvalidErr{}
	}
	return c.stdin.Write(b)
}

func (c *CmdConn) Close() error {
	c.stdin.Close()
	return c.cmd.Wait()
}

func (c *CmdConn) ok() bool {
	return c != nil && c.cmd != nil && c.stdin != nil && c.stdout != nil && c.stderr != nil
}
