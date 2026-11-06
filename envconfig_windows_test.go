//go:build windows

package envconfig

import (
	"os"
	"strings"
	"testing"
)

// TestCheckDisallowedCaseInsensitive covers issue #236: Process finds
// lowercase-prefixed variables via LookupEnv on Windows, but CheckDisallowed
// previously used a case-sensitive prefix scan of os.Environ() and missed them.
func TestCheckDisallowedCaseInsensitive(t *testing.T) {
	var s struct{ Known string }

	os.Clearenv()
	os.Setenv("app_unknown", "x")
	err := CheckDisallowed("app", &s)
	if err == nil {
		t.Fatal("unknown environment variable was not reported")
	}
	if !strings.Contains(err.Error(), "app_unknown") && !strings.Contains(err.Error(), "APP_UNKNOWN") {
		t.Fatalf("unexpected error: %v", err)
	}

	os.Clearenv()
	os.Setenv("app_known", "value")
	if err := CheckDisallowed("app", &s); err != nil {
		t.Fatalf("known lowercase key should be allowed: %v", err)
	}

	os.Clearenv()
	os.Setenv("APP_UNKNOWN", "x")
	err = CheckDisallowed("app", &s)
	if err == nil {
		t.Fatal("uppercase unknown variable was not reported")
	}
}
