// Copyright (c) 2016 Kelsey Hightower and others. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
)

const (
	// NoHeader constant to use for no header on usage output
	NoHeader = ""

	// DefaultHeader constant to use for default header on usage output
	DefaultHeader = `USAGE: {{ . }}
  This application is configured via the environment. The following environment
  variables can used specified:

`
	// DefaultListFormat constant to use to display usage in a list format
	DefaultListFormat = `  {{.Key}}
    [description] {{.Description}}
    [type]        {{.Type}}
    [default]     {{.Default}}
    [required]    {{.Required}}`

	// DefaultTableFormat constant to use to display usage in a tabluar format
	DefaultTableFormat = "table  {{.Key}}\t{{.Type}}\t{{.Default}}\t{{.Required}}\t{{.Description}}"
)

// usageInfo used to provide values for tabular output to the template function
type usageInfo struct {
	Key         string
	Description string
	Type        string
	Default     string
	Required    string
}

// toTypeDescription convert techie type information into something more human
func toTypeDescription(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		return fmt.Sprintf("List of %s", toTypeDescription(t.Elem()))
	case reflect.Ptr:
		return fmt.Sprintf("Reference to %s", toTypeDescription(t.Elem()))
	case reflect.Chan:
		return fmt.Sprintf("Channel of %s", toTypeDescription(t.Elem()))
	case reflect.Map:
		return fmt.Sprintf("Map of %s to %s", toTypeDescription(t.Key()), toTypeDescription(t.Elem()))
	case reflect.Func:
		return "Function"
	case reflect.Interface:
		return "Interface"
	case reflect.Struct:
		if t.Name() != "" {
			return t.Name()
		}
		return "Nested Struct"
	case reflect.String:
		name := t.Name()
		if name != "" && name != "string" {
			return name
		}
		return "String"
	case reflect.Bool:
		name := t.Name()
		if name != "" && name != "bool" {
			return name
		}
		return "Boolean"
	case reflect.Int:
		name := t.Name()
		if name != "" && name != "int" {
			return name
		}
		return "Integer"
	case reflect.Int8:
		name := t.Name()
		if name != "" && name != "int8" {
			return name
		}
		return "Integer(8 bits)"
	case reflect.Int16:
		name := t.Name()
		if name != "" && name != "int16" {
			return name
		}
		return "Integer(16 bits)"
	case reflect.Int32:
		name := t.Name()
		if name != "" && name != "int32" {
			return name
		}
		return "Integer(32 bits}"
	case reflect.Int64:
		name := t.Name()
		if name != "" && name != "int64" {
			return name
		}
		return "Integer(64 bits)"
	case reflect.Uint:
		name := t.Name()
		if name != "" && name != "uint" {
			return name
		}
		return "Unsigned Integer"
	case reflect.Uint8:
		name := t.Name()
		if name != "" && name != "uint8" {
			return name
		}
		return "Unsigned Integer(8 bits)"
	case reflect.Uint16:
		name := t.Name()
		if name != "" && name != "uint16" {
			return name
		}
		return "Unsigned Integer(16 bits)"
	case reflect.Uint32:
		name := t.Name()
		if name != "" && name != "uint32" {
			return name
		}
		return "Unsigned Integer(32 bits}"
	case reflect.Uint64:
		name := t.Name()
		if name != "" && name != "uint64" {
			return name
		}
		return "Unsigned Integer(64 bits)"
	case reflect.Float32:
		name := t.Name()
		if name != "" && name != "float32" {
			return name
		}
		return "Float"
	case reflect.Float64:
		name := t.Name()
		if name != "" && name != "float64" {
			return name
		}
		return "Float(64 bits)"
	case reflect.Complex64:
		return "Complex(64 bits)"
	case reflect.Complex128:
		return "Complex(128 bits)"
	}
	return fmt.Sprintf("%+v", t)
}

// Usage writes usage information to stderr using the default header and table format
func Usage(prefix string, spec interface{}) error {
	return Usagef(prefix, spec, os.Stderr, DefaultHeader, DefaultTableFormat)
}

// Usagef writes usage information to the specified io.Writer using the specifed header and usage format
func Usagef(prefix string, spec interface{}, out io.Writer, header string, format string) error {

	// gather first
	infos, err := GatherInfo(prefix, spec)
	if err != nil {
		return err
	}

	if header != NoHeader {
		headerTmpl, err := template.New("envconfig_header").Parse(header)
		if err != nil {
			return err
		}
		err = headerTmpl.Execute(out, path.Base(os.Args[0]))
		if err != nil {
			return err
		}
	}

	var tabs *tabwriter.Writer = nil
	var tmpl *template.Template
	var tmplSpec = format
	var usage usageInfo

	// If format is prefixed with "table" then strip it off to get the per-line template, display the table headers,
	// and inject a tab write filter
	if strings.HasPrefix(format, "table") {
		tmplSpec = strings.TrimPrefix(format, "table")
		tmpl, err = template.New("envconfig").Parse(tmplSpec)
		if err != nil {
			return nil
		}
		tabs = tabwriter.NewWriter(out, 1, 0, 4, ' ', 0)
		out = tabs

		usage = usageInfo{
			Key:         "KEY",
			Description: "DESCRIPTION",
			Type:        "TYPE",
			Default:     "DEFAULT",
			Required:    "REQUIRED",
		}
		if err = tmpl.Execute(out, usage); err != nil {
			return err
		}
		fmt.Fprintln(out)
	} else {
		// Not a table, so just use given filter as is
		tmpl, err = template.New("envconfig").Parse(tmplSpec)
		if err != nil {
			return err
		}
	}

	for _, info := range infos {
		req := info.Tags.Get("required")
		if req != "" {
			reqB, err := strconv.ParseBool(req)
			if err != nil {
				return err
			}
			if reqB {
				req = "true"
			}
		}
		usage = usageInfo{
			Key:         info.Key,
			Description: info.Tags.Get("desc"),
			Type:        toTypeDescription(info.Field.Type()),
			Default:     info.Tags.Get("default"),
			Required:    req,
		}

		if err := tmpl.Execute(out, usage); err != nil {
			return err
		}
		fmt.Fprintln(out)
	}
	// If we injected a tab writer then we need to flush
	if tabs != nil {
		tabs.Flush()
	}

	return nil
}
