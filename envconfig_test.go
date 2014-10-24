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
	Nested                       NestedSpecification
	NestedWithTag                NestedSpecification `envconfig:"ANOTHER_NESTED"`
}

type NestedSpecification struct {
	SomeValue    string
	ValueWithTag string `envconfig:"ANOTHER_VALUE"`
	Cool         bool
}

func TestProcess(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_RATE", "0.5")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	err := Process("env_config", &s)
	if err != nil {
		t.Error(err.Error())
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
}

func TestParseErrorBool(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "string")
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

func TestNestedSpecification(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_NESTED_SOMEVALUE", "string1")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.Nested.SomeValue != "string1" {
		t.Errorf("expected %q, got %q", "string1", s.Nested.SomeValue)
	}
}

func TestNestedSpecificationError(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_NESTED_COOL", "string")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Cool" {
		t.Errorf("expected %s, got %v", "Cool", v.FieldName)
	}
	if s.Nested.Cool != false {
		t.Errorf("expected %v, got %v", false, s.Nested.Cool)
	}
}

func TestNestedSpecificationWithTag(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_ANOTHER_NESTED_SOMEVALUE", "string2")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.NestedWithTag.SomeValue != "string2" {
		t.Errorf("expected %q, got %q", "string2", s.NestedWithTag.SomeValue)
	}
}

func TestNestedSpecificationWithTagOnProperty(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_NESTED_ANOTHER_VALUE", "string3")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.Nested.ValueWithTag != "string3" {
		t.Errorf("expected %q, got %q", "string3", s.Nested.ValueWithTag)
	}
}
