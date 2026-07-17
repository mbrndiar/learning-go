package sqlite

import (
	"net/url"
	"testing"
)

func TestSQLiteDSNWindowsDrivePathHasNoAuthority(t *testing.T) {
	dsn := sqliteDSN(`C:/Users/Go Learner/tasks #1?.db`)
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "file" {
		t.Fatalf("scheme = %q; want file", parsed.Scheme)
	}
	if parsed.Host != "" {
		t.Fatalf("host = %q; want empty", parsed.Host)
	}
	if parsed.Path != `/C:/Users/Go Learner/tasks #1?.db` {
		t.Fatalf("path = %q; want Windows drive path", parsed.Path)
	}
	if parsed.Query().Get("_pragma") != "busy_timeout(5000)" {
		t.Fatalf("_pragma = %q; want busy_timeout(5000)", parsed.Query().Get("_pragma"))
	}
}
