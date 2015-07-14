// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"os"
	"testing"
)

type Specification struct {
	Debug                        bool
	Port                         int
	Rate                         float32
	User                         string
	MultiWordVar                 string
	MultiWordVarWithAlt          string `envconfig:"MULTI_WORD_VAR_WITH_ALT"`
	MultiWordVarWithLowerCaseAlt string `envconfig:"multi_word_var_with_lower_case_alt"`
	NoPrefixWithAlt              string `envconfig:"SERVICE_HOST"`
	DefaultVar                   string `default:"foobar"`
	RequiredVar                  string `required:"true"`
	NoPrefixDefault              string `envconfig:"BROKER" default:"127.0.0.1"`
	RequiredDefault              string `required:"true" default:"foo2bar"`
}

func TestProcess(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_RATE", "0.5")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	os.Setenv("SERVICE_HOST", "127.0.0.1")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	if err != nil {
		t.Error(err.Error())
	}
	if s.NoPrefixWithAlt != "127.0.0.1" {
		t.Errorf("expected %v, got %v", "127.0.0.1", s.NoPrefixWithAlt)
	}
	if !s.Debug {
		t.Errorf("expected %v, got %v", true, s.Debug)
	}
	if s.Port != 8080 {
		t.Errorf("expected %d, got %v", 8080, s.Port)
	}
	if s.Rate != 0.5 {
		t.Errorf("expected %f, got %v", 0.5, s.Rate)
	}
	if s.User != "Kelsey" {
		t.Errorf("expected %s, got %s", "Kelsey", s.User)
	}
	if s.RequiredVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.RequiredVar)
	}
}

func TestParseErrorBool(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "string")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Debug" {
		t.Errorf("expected %s, got %v", "Debug", v.FieldName)
	}
	if s.Debug != false {
		t.Errorf("expected %v, got %v", false, s.Debug)
	}
}

func TestParseErrorFloat32(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_RATE", "string")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Rate" {
		t.Errorf("expected %s, got %v", "Rate", v.FieldName)
	}
	if s.Rate != 0 {
		t.Errorf("expected %v, got %v", 0, s.Rate)
	}
}

func TestParseErrorInt(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_PORT", "string")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Port" {
		t.Errorf("expected %s, got %v", "Port", v.FieldName)
	}
	if s.Port != 0 {
		t.Errorf("expected %v, got %v", 0, s.Port)
	}
}

func TestErrInvalidSpecification(t *testing.T) {
	m := make(map[string]string)
	err := Process("env_config", &m)
	if err != ErrInvalidSpecification {
		t.Errorf("expected %v, got %v", ErrInvalidSpecification, err)
	}
}

func TestUnsetVars(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("USER", "foo")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	// If the var is not defined the non-prefixed version should not be used
	// unless the struct tag says so
	if s.User != "" {
		t.Errorf("expected %q, got %q", "", s.User)
	}
}

func TestAlternateVarNames(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR", "foo")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT", "bar")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT", "baz")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	// Setting the alt version of the var in the environment has no effect if
	// the struct tag is not supplied
	if s.MultiWordVar != "" {
		t.Errorf("expected %q, got %q", "", s.MultiWordVar)
	}

	// Setting the alt version of the var in the environment correctly sets
	// the value if the struct tag IS supplied
	if s.MultiWordVarWithAlt != "bar" {
		t.Errorf("expected %q, got %q", "bar", s.MultiWordVarWithAlt)
	}

	// Alt value is not case sensitive and is treated as all uppercase
	if s.MultiWordVarWithLowerCaseAlt != "baz" {
		t.Errorf("expected %q, got %q", "baz", s.MultiWordVarWithLowerCaseAlt)
	}
}

func TestRequiredVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foobar")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.RequiredVar != "foobar" {
		t.Errorf("expected %s, got %s", "foobar", s.RequiredVar)
	}
}

func TestBlankDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "requiredvalue")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.DefaultVar != "foobar" {
		t.Errorf("expected %s, got %s", "foobar", s.DefaultVar)
	}
}

func TestNonBlankDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEFAULTVAR", "nondefaultval")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "requiredvalue")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.DefaultVar != "nondefaultval" {
		t.Errorf("expected %s, got %s", "nondefaultval", s.DefaultVar)
	}
}

func TestAlternateNameDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("BROKER", "betterbroker")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.NoPrefixDefault != "betterbroker" {
		t.Errorf("expected %q, got %q", "betterbroker", s.NoPrefixDefault)
	}

	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.NoPrefixDefault != "127.0.0.1" {
		t.Errorf("expected %q, got %q", "127.0.0.1", s.NoPrefixDefault)
	}
}

func TestRequiredDefault(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.RequiredDefault != "foo2bar" {
		t.Errorf("expected %q, got %q", "foo2bar", s.RequiredDefault)
	}
}

func TestMustProcess(t *testing.T) {

	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_RATE", "0.5")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	os.Setenv("SERVICE_HOST", "127.0.0.1")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	MustProcess("env_config", &s)

	defer func() {
		if err := recover(); err != nil {
			return
		}

		t.Error("expected panic")
	}()
	m := make(map[string]string)
	MustProcess("env_config", &m)
}
