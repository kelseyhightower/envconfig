package envconfig

import (
	"os"
	"testing"
)

type Specification struct {
	Debug    bool
	Password string
	Port     int
	User     string
}

func TestProcess(t *testing.T) {
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	var s Specification
	err := Process("env_config", &s)
	if err != nil {
		t.Error(err.Error())
	}
	if s.User != "Kelsey" {
		t.Errorf("expected %s, got %s", "Kelsey", s.User)
	}
	if s.Port != 8080 {
		t.Errorf("expected %d, got %v", 8080, s.Port)
	}
	if !s.Debug {
		t.Errorf("expected %v, got %v", true, s.Debug)
	}
}

func TestInvalidSpecificationError(t *testing.T) {
	m := make(map[string]string)
	err := Process("env_config", &m)
	if v, ok := err.(*InvalidSpecificationError); !ok {
		t.Errorf("expected InvalidSpecificationError, got %v", v)
	}
}
