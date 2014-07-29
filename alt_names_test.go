package envconfig

import (
	"os"
	"testing"
)

func TestAlternateVarNames(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR", "foo")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT", "bar")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT", "baz")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	// Setting the alt version of the var in the environment has no effect if
	// the alt tag is not supplied
	if s.MultiWordVar != "" {
		t.Errorf("expected %q, got %q", "", s.MultiWordVar)
	}

	// Setting the alt version of the var in the environment correctly sets
	// the value if the alt tag IS supplied
	if s.MultiWordVarWithAlt != "bar" {
		t.Errorf("expected %q, got %q", "bar", s.MultiWordVarWithAlt)
	}

	// Alt value is not case sensitive and is treated as all uppercase
	if s.MultiWordVarWithLowerCaseAlt != "baz" {
		t.Errorf("expected %q, got %q", "baz", s.MultiWordVarWithLowerCaseAlt)
	}
}

func TestAcceptSmushyName(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_ACCEPTSMUSHYNAME", "foo")
	os.Setenv("ENV_CONFIG_MULTIWORDVARWITHALT", "bar")
	os.Setenv("ENV_CONFIG_THIS_ONE_TOO", "baz")
	os.Setenv("ENV_CONFIG_THISONETOO", "bogus")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	// Smushy name is accepted when `accept_smushy_name` is specified and
	// non-smushy env var is not set
	if s.AcceptSmushyName != "foo" {
		t.Errorf("expected %q, got %q", "foo", s.AcceptSmushyName)
	}

	// Smushy name is not accepted on vars with an alt name specified and no
	// `accept_smusmy_name`
	if s.MultiWordVarWithAlt == "bar" {
		t.Errorf("did not expect %q, got %q", "bar", s.MultiWordVarWithAlt)
	}

	// Smushy name is only used as the default and is not accepted if the alt
	// env var is provided
	//
	// Also, yes this condition is logically redundant, but it helps explain
	// the use case
	if s.ThisOneToo != "baz" || s.ThisOneToo == "bogus" {
		t.Errorf("expected %q, got %q", "baz", s.ThisOneToo)
	}
}
