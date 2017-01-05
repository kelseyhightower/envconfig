// Copyright (c) 2016 Kelsey Hightower and others. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"text/tabwriter"
)

var TestUsageDefaultResult string = `This.application.is.configured.via.the.environment..The.following.environment
variables.can.used.specified:

KEY..............................................TYPE...............................DEFAULT...........REQUIRED....DESCRIPTION
ENV_CONFIG_EMBEDDED..............................Embedded.........................................................can.we.document.a.struct
ENV_CONFIG_ENABLED...............................True.or.False....................................................some.embedded.value
ENV_CONFIG_EMBEDDEDPORT..........................Integer..........................................................
ENV_CONFIG_MULTIWORDVAR..........................String...........................................................
ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT..............String...........................................................
ENV_CONFIG_EMBEDDED_WITH_ALT.....................String...........................................................
ENV_CONFIG_DEBUG.................................True.or.False....................................................
ENV_CONFIG_PORT..................................Integer..........................................................
ENV_CONFIG_RATE..................................Float............................................................
ENV_CONFIG_USER..................................String...........................................................
ENV_CONFIG_TTL...................................Unsigned.Integer.................................................
ENV_CONFIG_TIMEOUT...............................Duration.........................................................
ENV_CONFIG_ADMINUSERS............................Comma-separated.list.of.String...................................
ENV_CONFIG_MAGICNUMBERS..........................Comma-separated.list.of.Integer..................................
ENV_CONFIG_MULTIWORDVAR..........................String...........................................................
ENV_CONFIG_MULTI_WORD_VAR_WITH_AUTO_SPLIT........Unsigned.Integer.................................................
ENV_CONFIG_SOMEPOINTER...........................String...........................................................
ENV_CONFIG_SOMEPOINTERWITHDEFAULT................String.............................foo2baz.......................foorbar.is.the.word
ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT...............String...........................................................what.alt
ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT....String...........................................................
ENV_CONFIG_SERVICE_HOST..........................String...........................................................
ENV_CONFIG_DEFAULTVAR............................String.............................foobar........................
ENV_CONFIG_REQUIREDVAR...........................String...............................................true........
ENV_CONFIG_BROKER................................String.............................127.0.0.1.....................
ENV_CONFIG_REQUIREDDEFAULT.......................String.............................foo2bar...........true........
ENV_CONFIG_OUTER.................................Nested.Struct....................................................
ENV_CONFIG_OUTER_INNER...........................String...........................................................
ENV_CONFIG_OUTER_PROPERTYWITHDEFAULT.............String.............................fuzzybydefault................
ENV_CONFIG_AFTERNESTED...........................String...........................................................
ENV_CONFIG_HONOR.................................HonorDecodeInStruct..............................................
ENV_CONFIG_DATETIME..............................Time.............................................................
`

var TestUsageListResult string = `This.application.is.configured.via.the.environment..The.following.environment
variables.can.used.specified:

ENV_CONFIG_EMBEDDED
..[description].can.we.document.a.struct
..[type]........Embedded
..[default].....
..[required]....
ENV_CONFIG_ENABLED
..[description].some.embedded.value
..[type]........True.or.False
..[default].....
..[required]....
ENV_CONFIG_EMBEDDEDPORT
..[description].
..[type]........Integer
..[default].....
..[required]....
ENV_CONFIG_MULTIWORDVAR
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_EMBEDDED_WITH_ALT
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_DEBUG
..[description].
..[type]........True.or.False
..[default].....
..[required]....
ENV_CONFIG_PORT
..[description].
..[type]........Integer
..[default].....
..[required]....
ENV_CONFIG_RATE
..[description].
..[type]........Float
..[default].....
..[required]....
ENV_CONFIG_USER
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_TTL
..[description].
..[type]........Unsigned.Integer
..[default].....
..[required]....
ENV_CONFIG_TIMEOUT
..[description].
..[type]........Duration
..[default].....
..[required]....
ENV_CONFIG_ADMINUSERS
..[description].
..[type]........Comma-separated.list.of.String
..[default].....
..[required]....
ENV_CONFIG_MAGICNUMBERS
..[description].
..[type]........Comma-separated.list.of.Integer
..[default].....
..[required]....
ENV_CONFIG_MULTIWORDVAR
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_MULTI_WORD_VAR_WITH_AUTO_SPLIT
..[description].
..[type]........Unsigned.Integer
..[default].....
..[required]....
ENV_CONFIG_SOMEPOINTER
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_SOMEPOINTERWITHDEFAULT
..[description].foorbar.is.the.word
..[type]........String
..[default].....foo2baz
..[required]....
ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT
..[description].what.alt
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_SERVICE_HOST
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_DEFAULTVAR
..[description].
..[type]........String
..[default].....foobar
..[required]....
ENV_CONFIG_REQUIREDVAR
..[description].
..[type]........String
..[default].....
..[required]....true
ENV_CONFIG_BROKER
..[description].
..[type]........String
..[default].....127.0.0.1
..[required]....
ENV_CONFIG_REQUIREDDEFAULT
..[description].
..[type]........String
..[default].....foo2bar
..[required]....true
ENV_CONFIG_OUTER
..[description].
..[type]........Nested.Struct
..[default].....
..[required]....
ENV_CONFIG_OUTER_INNER
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_OUTER_PROPERTYWITHDEFAULT
..[description].
..[type]........String
..[default].....fuzzybydefault
..[required]....
ENV_CONFIG_AFTERNESTED
..[description].
..[type]........String
..[default].....
..[required]....
ENV_CONFIG_HONOR
..[description].
..[type]........HonorDecodeInStruct
..[default].....
..[required]....
ENV_CONFIG_DATETIME
..[description].
..[type]........Time
..[default].....
..[required]....
`

var TestUsageCustomResult = `ENV_CONFIG_EMBEDDED=can.we.document.a.struct
ENV_CONFIG_ENABLED=some.embedded.value
ENV_CONFIG_EMBEDDEDPORT=
ENV_CONFIG_MULTIWORDVAR=
ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT=
ENV_CONFIG_EMBEDDED_WITH_ALT=
ENV_CONFIG_DEBUG=
ENV_CONFIG_PORT=
ENV_CONFIG_RATE=
ENV_CONFIG_USER=
ENV_CONFIG_TTL=
ENV_CONFIG_TIMEOUT=
ENV_CONFIG_ADMINUSERS=
ENV_CONFIG_MAGICNUMBERS=
ENV_CONFIG_MULTIWORDVAR=
ENV_CONFIG_MULTI_WORD_VAR_WITH_AUTO_SPLIT=
ENV_CONFIG_SOMEPOINTER=
ENV_CONFIG_SOMEPOINTERWITHDEFAULT=foorbar.is.the.word
ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT=what.alt
ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT=
ENV_CONFIG_SERVICE_HOST=
ENV_CONFIG_DEFAULTVAR=
ENV_CONFIG_REQUIREDVAR=
ENV_CONFIG_BROKER=
ENV_CONFIG_REQUIREDDEFAULT=
ENV_CONFIG_OUTER=
ENV_CONFIG_OUTER_INNER=
ENV_CONFIG_OUTER_PROPERTYWITHDEFAULT=
ENV_CONFIG_AFTERNESTED=
ENV_CONFIG_HONOR=
ENV_CONFIG_DATETIME=
`

var TestUsageBadFormatResult = `{.Key}
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
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
{.Key}
`

func compareUsage(want, got string, t *testing.T) {
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

func TestUsageDefault(t *testing.T) {
	var s Specification
	os.Clearenv()
	save := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := Usage("env_config", &s)
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()
	w.Close()
	os.Stdout = save // restoring the real stdout
	out := <-outC

	if err != nil {
		t.Error(err.Error())
	}
	compareUsage(TestUsageDefaultResult, out, t)
}

func TestUsageTable(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	tabs := tabwriter.NewWriter(buf, 1, 0, 4, ' ', 0)
	err := Usagef("env_config", &s, tabs, DefaultTableFormat)
	tabs.Flush()
	if err != nil {
		t.Error(err.Error())
	}
	compareUsage(TestUsageDefaultResult, buf.String(), t)
}

func TestUsageList(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := Usagef("env_config", &s, buf, DefaultListFormat)
	if err != nil {
		t.Error(err.Error())
	}
	compareUsage(TestUsageListResult, buf.String(), t)
}

func TestUsageCustomFormat(t *testing.T) {
	var s Specification
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := Usagef("env_config", &s, buf, "{{range .}}{{usage_key .}}={{usage_description .}}\n{{end}}")
	if err != nil {
		t.Error(err.Error())
	}
	compareUsage(TestUsageCustomResult, buf.String(), t)
}

func TestUsageUnknownKeyFormat(t *testing.T) {
	var s Specification
	unknownError := "template: envconfig:1:2: executing \"envconfig\" at <.UnknownKey>"
	os.Clearenv()
	buf := new(bytes.Buffer)
	err := Usagef("env_config", &s, buf, "{{.UnknownKey}}")
	if err == nil {
		t.Errorf("expected 'unknown key' error, but got no error")
	}
	if strings.Index(err.Error(), unknownError) == -1 {
		t.Errorf("expected '%s', but got '%s'", unknownError, err.Error())
	}
}

func TestUsageBadFormat(t *testing.T) {
	var s Specification
	os.Clearenv()
	// If you don't use two {{}} then you get a lieteral
	buf := new(bytes.Buffer)
	err := Usagef("env_config", &s, buf, "{{range .}}{.Key}\n{{end}}")
	if err != nil {
		t.Error(err.Error())
	}
	compareUsage(TestUsageBadFormatResult, buf.String(), t)
}
