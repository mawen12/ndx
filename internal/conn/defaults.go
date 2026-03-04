package conn

import "os/user"

func defaultSettings() map[string]string {
	settings := make(map[string]string)

	settings["host"] = defaultHost()
	settings["port"] = "22"

	if u, err := user.Current(); err == nil {
		settings["user"] = u.Username
	}

	settings["command"] = "/bin/sh"

	return settings
}

func defaultHost() string {
	return "localhost"
}
