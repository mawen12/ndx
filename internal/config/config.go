package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func ParseConns(conns string) (QueryConns, error) {
	parts := strings.Split(conns, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("conns(%s) is invalid", conns)
	}

	var queryConns QueryConns
	//queryConns := make([]*QueryConn, len(parts))
	for _, part := range parts {
		qc, err := parseConn(part)
		if err != nil {
			return nil, err
		}

		queryConns = append(queryConns, qc)
	}

	return queryConns, nil
}

func parseConn(conn string) (*QueryConn, error) {
	defaultSettings := defaultSettings()

	connStringSettings := make(map[string]string)
	if strings.HasPrefix(conn, "cmd://") || strings.HasPrefix(conn, "ssh://") {
		var err error
		connStringSettings, err = parseURLSettings(conn)
		if err != nil {
			return nil, &parseConfigError{conns: conn, msg: "failed to parse as URL", err: err}
		}
	}

	settings := mergeSettings(defaultSettings, connStringSettings)

	qc := QueryConn{
		Origin:   conn,
		Scheme:   settings["scheme"],
		Host:     settings["host"],
		User:     settings["user"],
		Password: settings["password"],
		LogFile:  settings["logfile"],
	}

	if settings["port"] != "" {
		port, err := strconv.ParseUint(settings["port"], 10, 16)
		if err != nil {
			return nil, err
		}
		qc.Port = uint16(port)
	}

	return &qc, nil
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
