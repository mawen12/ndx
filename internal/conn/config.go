package conn

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/proto"

	"github.com/google/uuid"
)

type GetSSLPasswordFunc func(ctx context.Context) string

type Config struct {
	Scheme        string
	Host          string
	Port          uint16
	User          string
	Password      string
	Command       string
	LogFile       string
	DialFunc      DialFunc
	BuildFrontend BuildFrontendFunc
	RuntimeParams map[string]string

	Fallbacks            []*FallbackConfig
	OnNotice             NoticeHandler
	PathPrefix           string
	createdByParseConfig bool
}

type FallbackConfig struct {
	Host string
	Port uint16
}

type ParseConfigOptions struct {
	GetSSLPassword GetSSLPasswordFunc
}

func ParseConfig(connString string) (*Config, error) {
	defaultSettings := defaultSettings()

	connStringSettings := make(map[string]string)
	if strings.HasPrefix(connString, "cmd://") || strings.HasPrefix(connString, "ssh://") {
		var err error
		connStringSettings, err = parseURLSettings(connString)
		if err != nil {
			return nil, &parseConfigError{connString: connString, msg: "failed to parse as URL", err: err}
		}
	}

	settings := mergeSettings(defaultSettings, connStringSettings)

	config := &Config{
		createdByParseConfig: true,
		Scheme:               settings["scheme"],
		Host:                 settings["host"],
		User:                 settings["user"],
		Password:             settings["password"],
		Command:              settings["command"],
		LogFile:              settings["logfile"],
		RuntimeParams:        make(map[string]string),
		BuildFrontend:        makeDefaultBuildFrontendFunc(),
		PathPrefix:           fmt.Sprintf("/tmp/ndx_%s", uuid.New().String()),
	}

	if settings["port"] != "" {
		port, err := strconv.ParseUint(settings["port"], 10, 16)
		if err != nil {
			return nil, err
		}
		config.Port = uint16(port)
	}

	config.DialFunc = makeDialFunc(config.Scheme, config.Host)

	return config, nil
}

func mergeSettings(settingSets ...map[string]string) map[string]string {
	settings := make(map[string]string)

	for _, s2 := range settingSets {
		for k, v := range s2 {
			settings[k] = v
		}
	}

	return settings
}

func parseURLSettings(connString string) (map[string]string, error) {
	settings := make(map[string]string)

	u, err := url.Parse(connString)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "cmd" && u.Scheme != "ssh" {
		return nil, fmt.Errorf("expect scheme 'cmd' or 'ssh', but is %s", u.Scheme)
	}
	settings["scheme"] = u.Scheme

	if u.User != nil {
		settings["user"] = u.User.Username()
		if password, present := u.User.Password(); present {
			settings["password"] = password
		}
	}

	var hosts []string
	var ports []string
	for _, host := range strings.Split(u.Host, ",") {
		if host == "" {
			continue
		}
		if isIPOnly(host) {
			hosts = append(hosts, strings.Trim(host, "[]"))
			continue
		}
		h, p, err := net.SplitHostPort(host)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port in '%s', err: %w", host, err)
		}
		if h != "" {
			hosts = append(hosts, h)
		}
		if p != "" {
			ports = append(ports, p)
		}
	}

	if len(hosts) > 0 {
		settings["host"] = strings.Join(hosts, ",")
	}
	if len(ports) > 0 {
		settings["port"] = strings.Join(ports, ",")
	}

	if u.Path != "" {
		settings["logfile"] = u.Path
	}

	for k, v := range u.Query() {
		settings[k] = v[0]
	}

	return settings, nil
}

func isIPOnly(host string) bool {
	return net.ParseIP(strings.Trim(host, "[]")) != nil || !strings.Contains(host, ":")
}

func makeDefaultBuildFrontendFunc() BuildFrontendFunc {
	return func(conn model.Conn, w io.Writer) Frontend {
		sr := NewStringReader(conn)
		frontend := proto.NewFrontend(sr, w)

		return frontend
	}
}

func makeDialFunc(scheme string, host string) DialFunc {
	switch scheme {
	case "cmd":
		return func(ctx context.Context, config Config) (model.Conn, error) {
			return NewCmdConnConfig(ctx, config)
		}
	case "ssh":
		return func(ctx context.Context, config Config) (model.Conn, error) {
			return NewShellConnConfig(ctx, config)
		}
	default:
		panic("BUG: unsupported scheme for makeDialFunc")
	}
}

func NewCmdConnConfig(ctx context.Context, config Config) (*CmdConn, error) {
	return NewCmdConn(ctx, config.Command)
}

func NewShellConnConfig(ctx context.Context, config Config) (*ShellConn, error) {
	return NewShellConn(ctx, fmt.Sprintf("%s:%d", config.Host, int(config.Port)), config.User, config.Password)
}
