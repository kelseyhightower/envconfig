// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

type Specification struct {
	Embedded                     `desc:"can we document a struct"`
	EmbeddedButIgnored           `ignored:"true"`
	Debug                        bool
	Port                         int
	Rate                         float32
	User                         string
	TTL                          uint32
	Timeout                      time.Duration
	AdminUsers                   []string
	MagicNumbers                 []int
	MultiWordVar                 string
	SomePointer                  *string
	SomePointerWithDefault       *string `default:"foo2baz" desc:"foorbar is the word"`
	MultiWordVarWithAlt          string  `envconfig:"MULTI_WORD_VAR_WITH_ALT" desc:"what alt"`
	MultiWordVarWithLowerCaseAlt string  `envconfig:"multi_word_var_with_lower_case_alt"`
	NoPrefixWithAlt              string  `envconfig:"SERVICE_HOST"`
	DefaultVar                   string  `default:"foobar"`
	RequiredVar                  string  `required:"true"`
	NoPrefixDefault              string  `envconfig:"BROKER" default:"127.0.0.1"`
	RequiredDefault              string  `required:"true" default:"foo2bar"`
	Ignored                      string  `ignored:"true"`
}

type Embedded struct {
	Enabled             bool `desc:"some embedded value"`
	EmbeddedPort        int
	MultiWordVar        string
	MultiWordVarWithAlt string `envconfig:"MULTI_WITH_DIFFERENT_ALT"`
	EmbeddedAlt         string `envconfig:"EMBEDDED_WITH_ALT"`
	EmbeddedIgnored     string `ignored:"true"`
}

type EmbeddedButIgnored struct {
	FirstEmbeddedButIgnored  string
	SecondEmbeddedButIgnored string
}

var TestDocumentDefaultResult string = `USAGE:.envconfig.test

..This.application.is.configured.via.the.environment..The.following.environment
..variables.can.used.specified:

..KEY..............................................TYPE.........................DEFAULT......REQUIRED....DESCRIPTION
..ENV_CONFIG_ENABLED...............................Boolean...................................false.......some.embedded.value
..ENV_CONFIG_EMBEDDEDPORT..........................Integer...................................false.......
..ENV_CONFIG_MULTIWORDVAR..........................String....................................false.......
..ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT..............String....................................false.......
..ENV_CONFIG_EMBEDDED_WITH_ALT.....................String....................................false.......
..ENV_CONFIG_EMBEDDED..............................Embedded..................................false.......can.we.document.a.struct
..ENV_CONFIG_DEBUG.................................Boolean...................................false.......
..ENV_CONFIG_PORT..................................Integer...................................false.......
..ENV_CONFIG_RATE..................................Float.....................................false.......
..ENV_CONFIG_USER..................................String....................................false.......
..ENV_CONFIG_TTL...................................Unsigned.Integer(32.bits}.................false.......
..ENV_CONFIG_TIMEOUT...............................Integer(64.bits)..........................false.......
..ENV_CONFIG_ADMINUSERS............................List.of.String............................false.......
..ENV_CONFIG_MAGICNUMBERS..........................List.of.Integer...........................false.......
..ENV_CONFIG_MULTIWORDVAR..........................String....................................false.......
..ENV_CONFIG_SOMEPOINTER...........................String....................................false.......
..ENV_CONFIG_SOMEPOINTERWITHDEFAULT................String.......................foo2baz......false.......foorbar.is.the.word
..ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT...............String....................................false.......what.alt
..ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT....String....................................false.......
..ENV_CONFIG_SERVICE_HOST..........................String....................................false.......
..ENV_CONFIG_DEFAULTVAR............................String.......................foobar.......false.......
..ENV_CONFIG_REQUIREDVAR...........................String....................................true........
..ENV_CONFIG_BROKER................................String.......................127.0.0.1....false.......
..ENV_CONFIG_REQUIREDDEFAULT.......................String.......................foo2bar......true........
`

var TestDocumentListResult string = `..ENV_CONFIG_ENABLED
....[description].some.embedded.value
....[type]........Boolean
....[default].....
....[required]....false
..ENV_CONFIG_EMBEDDEDPORT
....[description].
....[type]........Integer
....[default].....
....[required]....false
..ENV_CONFIG_MULTIWORDVAR
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_EMBEDDED_WITH_ALT
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_EMBEDDED
....[description].can.we.document.a.struct
....[type]........Embedded
....[default].....
....[required]....false
..ENV_CONFIG_DEBUG
....[description].
....[type]........Boolean
....[default].....
....[required]....false
..ENV_CONFIG_PORT
....[description].
....[type]........Integer
....[default].....
....[required]....false
..ENV_CONFIG_RATE
....[description].
....[type]........Float
....[default].....
....[required]....false
..ENV_CONFIG_USER
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_TTL
....[description].
....[type]........Unsigned.Integer(32.bits}
....[default].....
....[required]....false
..ENV_CONFIG_TIMEOUT
....[description].
....[type]........Integer(64.bits)
....[default].....
....[required]....false
..ENV_CONFIG_ADMINUSERS
....[description].
....[type]........List.of.String
....[default].....
....[required]....false
..ENV_CONFIG_MAGICNUMBERS
....[description].
....[type]........List.of.Integer
....[default].....
....[required]....false
..ENV_CONFIG_MULTIWORDVAR
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_SOMEPOINTER
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_SOMEPOINTERWITHDEFAULT
....[description].foorbar.is.the.word
....[type]........String
....[default].....foo2baz
....[required]....false
..ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT
....[description].what.alt
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_SERVICE_HOST
....[description].
....[type]........String
....[default].....
....[required]....false
..ENV_CONFIG_DEFAULTVAR
....[description].
....[type]........String
....[default].....foobar
....[required]....false
..ENV_CONFIG_REQUIREDVAR
....[description].
....[type]........String
....[default].....
....[required]....true
..ENV_CONFIG_BROKER
....[description].
....[type]........String
....[default].....127.0.0.1
....[required]....false
..ENV_CONFIG_REQUIREDDEFAULT
....[description].
....[type]........String
....[default].....foo2bar
....[required]....true
`

var TestDocumentCustomResult = `ENV_CONFIG_ENABLED=some.embedded.value
ENV_CONFIG_EMBEDDEDPORT=
ENV_CONFIG_MULTIWORDVAR=
ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT=
ENV_CONFIG_EMBEDDED_WITH_ALT=
ENV_CONFIG_EMBEDDED=can.we.document.a.struct
ENV_CONFIG_DEBUG=
ENV_CONFIG_PORT=
ENV_CONFIG_RATE=
ENV_CONFIG_USER=
ENV_CONFIG_TTL=
ENV_CONFIG_TIMEOUT=
ENV_CONFIG_ADMINUSERS=
ENV_CONFIG_MAGICNUMBERS=
ENV_CONFIG_MULTIWORDVAR=
ENV_CONFIG_SOMEPOINTER=
ENV_CONFIG_SOMEPOINTERWITHDEFAULT=foorbar.is.the.word
ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT=what.alt
ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT=
ENV_CONFIG_SERVICE_HOST=
ENV_CONFIG_DEFAULTVAR=
ENV_CONFIG_REQUIREDVAR=
ENV_CONFIG_BROKER=
ENV_CONFIG_REQUIREDDEFAULT=
`

var TestDocumentBadFormatResult = `{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
`

func compareDocs(want, got string, t *testing.T) {
	have := strings.Replace(got, " ", ".", -1)
	if want != have {
		shortest := len(want)
		if len(have) < shortest {
			shortest = len(have)
		}
		if len(want) != len(have) {
			t.Errorf("expected result length of %d, found %d", len(want), len(have))
		}
		for i := 0; i < shortest; i++ {
			if want[i] != have[i] {
				t.Errorf("difference at index %d, expected '%c' (%v), found '%c' (%v)\n",
					i, want[i], want[i], have[i], have[i])
				break
			}
		}
		t.Errorf("Complete Expected:\n'%s'\nComplete Found:\n'%s'\n", want, have)
	}
}

func TestDocumentDefault(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := Document("env_config", &s, buf)
	if err != nil {
		t.Error(err.Error())
	}
	compareDocs(TestDocumentDefaultResult, buf.String(), t)
}

func TestDocumentWithoutHeader(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := DocumentFormat("env_config", &s, buf, NoHeader, DefaultTableFormat)
	if err != nil {
		t.Error(err.Error())
	}
	compareDocs(TestDocumentDefaultResult[136:], buf.String(), t)
}

func TestDocumentList(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := DocumentFormat("env_config", &s, buf, NoHeader, DefaultListFormat)
	if err != nil {
		t.Error(err.Error())
	}
	compareDocs(TestDocumentListResult, buf.String(), t)
}

func TestDocumentCustomFormat(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := DocumentFormat("env_config", &s, buf, NoHeader, "{{.Key}}={{.Description}}")
	if err != nil {
		t.Error(err.Error())
	}
	compareDocs(TestDocumentCustomResult, buf.String(), t)
}

func TestDocumentUnknownKeyFormat(t *testing.T) {
	var s Specification
	unknownError := "template: envconfig:1:2: executing \"envconfig\" at <.UnknownKey>"
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := DocumentFormat("env_config", &s, buf, NoHeader, "{{.UnknownKey}}")
	if err == nil {
		t.Errorf("expected 'unknown key' error, but got no error")
	}
	if err.Error()[:len(unknownError)] != unknownError {
		t.Errorf("expected '%s', but got '%s'", unknownError, err.Error())
	}
}

func TestDocumentBadFormat(t *testing.T) {
	var s Specification
	os.Clearenv()
	// If you don't use two {{}} then you get a lieteral
	buf := new(bytes.Buffer)
	err := DocumentFormat("env_config", &s, buf, NoHeader, "{.Key}")
	if err != nil {
		t.Error(err.Error())
	}
	compareDocs(TestDocumentBadFormatResult, buf.String(), t)
}

func TestProcess(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_RATE", "0.5")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	os.Setenv("ENV_CONFIG_TIMEOUT", "2m")
	os.Setenv("ENV_CONFIG_ADMINUSERS", "John,Adam,Will")
	os.Setenv("ENV_CONFIG_MAGICNUMBERS", "5,10,20")
	os.Setenv("SERVICE_HOST", "127.0.0.1")
	os.Setenv("ENV_CONFIG_TTL", "30")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	os.Setenv("ENV_CONFIG_IGNORED", "was-not-ignored")
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
	if s.TTL != 30 {
		t.Errorf("expected %d, got %v", 30, s.TTL)
	}
	if s.User != "Kelsey" {
		t.Errorf("expected %s, got %s", "Kelsey", s.User)
	}
	if s.Timeout != 2*time.Minute {
		t.Errorf("expected %s, got %s", 2*time.Minute, s.Timeout)
	}
	if s.RequiredVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.RequiredVar)
	}
	if len(s.AdminUsers) != 3 ||
		s.AdminUsers[0] != "John" ||
		s.AdminUsers[1] != "Adam" ||
		s.AdminUsers[2] != "Will" {
		t.Errorf("expected %#v, got %#v", []string{"John", "Adam", "Will"}, s.AdminUsers)
	}
	if len(s.MagicNumbers) != 3 ||
		s.MagicNumbers[0] != 5 ||
		s.MagicNumbers[1] != 10 ||
		s.MagicNumbers[2] != 20 {
		t.Errorf("expected %#v, got %#v", []int{5, 10, 20}, s.MagicNumbers)
	}
	if s.Ignored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
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

func TestParseErrorUint(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_TTL", "-30")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "TTL" {
		t.Errorf("expected %s, got %v", "TTL", v.FieldName)
	}
	if s.TTL != 0 {
		t.Errorf("expected %v, got %v", 0, s.TTL)
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

	if *s.SomePointerWithDefault != "foo2baz" {
		t.Errorf("expected %s, got %s", "foo2baz", *s.SomePointerWithDefault)
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

func TestExplicitBlankDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEFAULTVAR", "")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "")

	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.DefaultVar != "" {
		t.Errorf("expected %s, got %s", "\"\"", s.DefaultVar)
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

func TestPointerFieldBlank(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.SomePointer != nil {
		t.Errorf("expected <nil>, got %2", *s.SomePointer)
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

func TestEmbeddedStruct(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "required")
	os.Setenv("ENV_CONFIG_ENABLED", "true")
	os.Setenv("ENV_CONFIG_EMBEDDEDPORT", "1234")
	os.Setenv("ENV_CONFIG_MULTIWORDVAR", "foo")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT", "bar")
	os.Setenv("ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT", "baz")
	os.Setenv("ENV_CONFIG_EMBEDDED_WITH_ALT", "foobar")
	os.Setenv("ENV_CONFIG_SOMEPOINTER", "foobaz")
	os.Setenv("ENV_CONFIG_EMBEDDED_IGNORED", "was-not-ignored")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}
	if !s.Enabled {
		t.Errorf("expected %v, got %v", true, s.Enabled)
	}
	if s.EmbeddedPort != 1234 {
		t.Errorf("expected %d, got %v", 1234, s.EmbeddedPort)
	}
	if s.MultiWordVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.MultiWordVar)
	}
	if s.Embedded.MultiWordVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.Embedded.MultiWordVar)
	}
	if s.MultiWordVarWithAlt != "bar" {
		t.Errorf("expected %s, got %s", "bar", s.MultiWordVarWithAlt)
	}
	if s.Embedded.MultiWordVarWithAlt != "baz" {
		t.Errorf("expected %s, got %s", "baz", s.Embedded.MultiWordVarWithAlt)
	}
	if s.EmbeddedAlt != "foobar" {
		t.Errorf("expected %s, got %s", "foobar", s.EmbeddedAlt)
	}
	if *s.SomePointer != "foobaz" {
		t.Errorf("expected %s, got %s", "foobaz", *s.SomePointer)
	}
	if s.EmbeddedIgnored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}
}

func TestEmbeddedButIgnoredStruct(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "required")
	os.Setenv("ENV_CONFIG_FIRSTEMBEDDEDBUTIGNORED", "was-not-ignored")
	os.Setenv("ENV_CONFIG_SECONDEMBEDDEDBUTIGNORED", "was-not-ignored")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}
	if s.FirstEmbeddedButIgnored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}
	if s.SecondEmbeddedButIgnored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}
}

func TestNonPointerFailsProperly(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "snap")

	err := Process("env_config", s)
	if err != ErrInvalidSpecification {
		t.Errorf("non-pointer should fail with ErrInvalidSpecification, was instead %s", err)
	}
}

func TestCustomDecoder(t *testing.T) {
	s := struct {
		Foo string
		Bar bracketed
	}{}

	os.Clearenv()
	os.Setenv("ENV_CONFIG_FOO", "foo")
	os.Setenv("ENV_CONFIG_BAR", "bar")

	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.Foo != "foo" {
		t.Errorf("foo: expected 'foo', got %q", s.Foo)
	}

	if string(s.Bar) != "[bar]" {
		t.Errorf("bar: expected '[bar]', got %q", string(s.Bar))
	}
}

func TestCustomDecoderWithPointer(t *testing.T) {
	s := struct {
		Foo string
		Bar *bracketed
	}{}

	// Decode would panic when b is nil, so make sure it
	// has an initial value to replace.
	var b bracketed = "initial_value"
	s.Bar = &b

	os.Clearenv()
	os.Setenv("ENV_CONFIG_FOO", "foo")
	os.Setenv("ENV_CONFIG_BAR", "bar")

	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.Foo != "foo" {
		t.Errorf("foo: expected 'foo', got %q", s.Foo)
	}

	if string(*s.Bar) != "[bar]" {
		t.Errorf("bar: expected '[bar]', got %q", string(*s.Bar))
	}
}

type bracketed string

func (b *bracketed) Decode(value string) error {
	*b = bracketed("[" + value + "]")
	return nil
}
