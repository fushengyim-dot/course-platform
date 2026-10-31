package config

import (
	"strings"
	"testing"
)

func good() map[string]string {
	return map[string]string{"APP_VERSION": "test", "HTTP_ADDR": ":8080", "MYSQL_HOST": "mysql", "MYSQL_PORT": "3306", "MYSQL_DATABASE": "test", "MYSQL_USER": "test", "MYSQL_PASSWORD": "test-only-password", "QDRANT_URL": "http://qdrant:6333"}
}
func TestNormalConfig(t *testing.T) {
	m := good()
	if _, e := Load(func(k string) string { return m[k] }); e != nil {
		t.Fatal(e)
	}
}
func TestFailureMissingEveryKey(t *testing.T) {
	for key := range good() {
		t.Run(key, func(t *testing.T) {
			m := good()
			delete(m, key)
			_, e := Load(func(k string) string { return m[k] })
			if e == nil || !strings.Contains(e.Error(), key) {
				t.Fatal("missing key accepted")
			}
			if strings.Contains(e.Error(), m["MYSQL_PASSWORD"]) && m["MYSQL_PASSWORD"] != "" {
				t.Fatal("secret leaked")
			}
		})
	}
}
func TestBoundaryInvalidConfig(t *testing.T) {
	for _, v := range []struct{ k, v string }{{"MYSQL_PORT", "0"}, {"MYSQL_PORT", "65536"}, {"HTTP_ADDR", ":abc"}, {"QDRANT_URL", "file:///secret"}, {"QDRANT_URL", "http://u:secret@host"}, {"QDRANT_URL", "http://host?secret=x"}, {"QDRANT_URL", "http://host#fragment"}, {"APP_VERSION", strings.Repeat("a", 65)}} {
		m := good()
		m[v.k] = v.v
		if _, e := Load(func(k string) string { return m[k] }); e == nil {
			t.Errorf("accepted invalid %s", v.k)
		}
	}
}
