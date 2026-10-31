package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type Config struct{ Version, HTTPAddr, MySQLHost, MySQLPort, MySQLDatabase, MySQLUser, MySQLPassword, QdrantURL string }

func Load(get func(string) string) (Config, error) {
	keys := []string{"APP_VERSION", "HTTP_ADDR", "MYSQL_HOST", "MYSQL_PORT", "MYSQL_DATABASE", "MYSQL_USER", "MYSQL_PASSWORD", "QDRANT_URL"}
	for _, k := range keys {
		if strings.TrimSpace(get(k)) == "" {
			return Config{}, fmt.Errorf("invalid configuration: %s", k)
		}
	}
	c := Config{get(keys[0]), get(keys[1]), get(keys[2]), get(keys[3]), get(keys[4]), get(keys[5]), get(keys[6]), get(keys[7])}
	if len(c.Version) > 64 || strings.ContainsAny(c.Version, "\r\n") {
		return Config{}, fmt.Errorf("invalid configuration: APP_VERSION")
	}
	_, port, e := net.SplitHostPort(c.HTTPAddr)
	if e != nil || !validPort(port) {
		return Config{}, fmt.Errorf("invalid configuration: HTTP_ADDR")
	}
	if !validPort(c.MySQLPort) {
		return Config{}, fmt.Errorf("invalid configuration: MYSQL_PORT")
	}
	u, e := url.Parse(c.QdrantURL)
	if e != nil || u.Hostname() == "" || (u.Port() != "" && !validPort(u.Port())) || strings.HasSuffix(u.Host, ":") || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return Config{}, fmt.Errorf("invalid configuration: QDRANT_URL")
	}
	return c, nil
}
func validPort(s string) bool { n, e := strconv.Atoi(s); return e == nil && n > 0 && n <= 65535 }
