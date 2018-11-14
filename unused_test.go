package envconfig

import (
	"os"
	"sort"
	"testing"
)

func TestUnused(t *testing.T) {
	spec := struct {
		Used   string
		Unset  string
		Nested struct {
			Used  string
			Unset string
		}
	}{}

	t.Run("with prefix", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_UNUSED_CONFIGS_USED", "set")
		os.Setenv("TEST_UNUSED_CONFIGS_UNUSED", "set")
		os.Setenv("UNRELATED_UNUSED", "set")
		os.Setenv("TEST_UNUSED_CONFIGS_NESTED_USED", "set")
		os.Setenv("TEST_UNUSED_CONFIGS_NESTED_UNUSED", "set")

		unused, err := Unused("test_unused_configs", &spec)
		if err != nil {
			t.Fatalf("Err shuold be nil, was: %s", err)
		}

		expectedUnused := []string{
			"TEST_UNUSED_CONFIGS_UNUSED",
			"TEST_UNUSED_CONFIGS_NESTED_UNUSED",
		}
		sort.Strings(unused)
		sort.Strings(expectedUnused)

		if !equal(expectedUnused, unused) {
			t.Errorf("Expected %+v as unused vars, got %+v", expectedUnused, unused)
		}
	})

	t.Run("without prefix", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("USED", "set")
		os.Setenv("UNUSED", "set")
		os.Setenv("UNRELATED", "set")
		os.Setenv("NESTED_USED", "set")
		os.Setenv("NESTED_UNUSED", "set")

		unused, err := Unused("", &spec)
		if err != nil {
			t.Fatalf("Err shuold be nil, was: %s", err)
		}

		expectedUnused := []string{
			"UNUSED",
			"NESTED_UNUSED",
			"UNRELATED",
		}
		sort.Strings(unused)
		sort.Strings(expectedUnused)

		if !equal(expectedUnused, unused) {
			t.Errorf("Expected %+v as unused vars, got %+v", expectedUnused, unused)
		}
	})
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}