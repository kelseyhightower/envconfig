// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"text/template"
	"time"
)

const (
	// True is true
	TrueString = "true"
	// False is false
	FalseString = "false"
)

// ErrInvalidSpecification indicates that a specification is of the wrong type.
var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

// DefaultListFormat public format that can be used to display options in a list
var DefaultListFormat string = `  {{.Key}}
    [description] {{.Description}}
    [type]        {{.Type}}
    [default]     {{.Default}}
    [required]    {{.Required}}`

// DefaultTableFormat public format that can be used to display options in a table, default format if not specified
var DefaultTableFormat string = "table  {{.Key}}\t{{.Type}}\t{{.Default}}\t{{.Required}}\t{{.Description}}"

// A ParseError occurs when an environment variable cannot be converted to
// the type required by a struct field during assignment.
type ParseError struct {
	KeyName   string
	FieldName string
	TypeName  string
	Value     string
}

// A Decoder is a type that knows how to de-serialize environment variables
// into itself.
type Decoder interface {
	Decode(value string) error
}

// ProcessFunc vistor function definition while traversing the list of options
type ProcessFunc func(field reflect.Value, tof reflect.StructField, fieldName string, key string, alt_key string, def string, required string) error

func (e *ParseError) Error() string {
	return fmt.Sprintf("envconfig.Process: assigning %[1]s to %[2]s: converting '%[3]s' to type %[4]s", e.KeyName, e.FieldName, e.Value, e.TypeName)
}

// Visit vists all the configuration options, calling the process function on each
func Visit(prefix string, spec interface{}, process ProcessFunc) error {
	s := reflect.ValueOf(spec)

	if s.Kind() != reflect.Ptr {
		return ErrInvalidSpecification
	}

	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return ErrInvalidSpecification
	}

	typeOfSpec := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		tof := typeOfSpec.Field(i)
		if !f.CanSet() || tof.Tag.Get("ignored") == TrueString {
			continue
		}

		if tof.Anonymous && f.Kind() == reflect.Struct {
			embeddedPtr := f.Addr().Interface()
			if err := Visit(prefix, embeddedPtr, process); err != nil {
				return err
			}
			f.Set(reflect.ValueOf(embeddedPtr).Elem())
		}

		alt := tof.Tag.Get("envconfig")
		fieldName := tof.Name
		if alt != "" {
			fieldName = alt
		}
		key := strings.ToUpper(fmt.Sprintf("%s_%s", prefix, fieldName))
		def := tof.Tag.Get("default")
		req := tof.Tag.Get("required")
		if req != TrueString {
			req = FalseString
		}

		if err := process(f, tof, fieldName, key, alt, def, req); err != nil {
			return err
		}
	}
	return nil
}

// DocumentInfo used to manage configuraiton option information for template output
type DocumentInfo struct {
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
		return toTypeDescription(t.Elem())
	case reflect.String:
		return "String"
	case reflect.Bool:
		return "Boolean"
	case reflect.Int:
		return "Integer"
	case reflect.Int8:
		return "Integer(8 bits)"
	case reflect.Int16:
		return "Integer(16 bits)"
	case reflect.Int32:
		return "Integer(32 bits}"
	case reflect.Int64:
		return "Integer(64 bits)"
	case reflect.Uint:
		return "Unsigned Integer"
	case reflect.Uint8:
		return "Unsigned Integer(8 bits)"
	case reflect.Uint16:
		return "Unsigned Integer(16 bits)"
	case reflect.Uint32:
		return "Unsigned Integer(32 bits}"
	case reflect.Uint64:
		return "Unsigned Integer(64 bits)"
	case reflect.Float32:
		return "Float"
	case reflect.Float64:
		return "Float(64 bits)"
	}
	return t.Name()
}

// Document write the default format of documentation for configuration options
func Document(prefix string, spec interface{}, out io.Writer) error {
	return DocumentFormat(prefix, spec, out, true, DefaultTableFormat)
}

// DocumentFormat write the configration options using the given template
func DocumentFormat(prefix string, spec interface{}, out io.Writer, showHeader bool, format string) error {
	var info DocumentInfo

	if showHeader {
		fmt.Fprintf(out, "USAGE: %s\n", path.Base(os.Args[0]))
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  This application is configured via the environment. The following environment")
		fmt.Fprintln(out, "  variables can used specified:")
		fmt.Fprintln(out)
	}

	var tabs *tabwriter.Writer = nil
	var tmpl *template.Template
	var tmplSpec = format
	var err error

	// If format is prefixed with "table" then strip it off to get the per-line template, display the table headers,
	// and inject a tab write filter
	if strings.HasPrefix(format, "table") {
		tmplSpec = strings.TrimPrefix(format, "table")
		tmpl, err = template.New("envconfig").Parse(tmplSpec)
		if err != nil {
			return err
		}
		tabs = tabwriter.NewWriter(out, 1, 0, 4, ' ', 0)
		out = tabs

		info = DocumentInfo{
			Key:         "KEY",
			Description: "DESCRIPTION",
			Type:        "TYPE",
			Default:     "DEFAULT",
			Required:    "REQUIRED",
		}
		err = tmpl.Execute(out, info)
		if err != nil {
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

	// Visit the configuration options and output a line for each option
	err = Visit(prefix, spec, func(field reflect.Value, tof reflect.StructField, fieldName string,
		key string, alt_key string, def string, req string) error {

		info = DocumentInfo{
			Key:         key,
			Description: tof.Tag.Get("desc"),
			Type:        toTypeDescription(tof.Type),
			Default:     def,
			Required:    req,
		}

		if err := tmpl.Execute(out, info); err != nil {
			return err
		}
		fmt.Fprintln(out)

		return nil
	})

	if err != nil {
		return err
	}

	// If we injected a tab writer then we need to flush
	if tabs != nil {
		tabs.Flush()
	}

	return nil

}

// Process populates the specfied struct based on the environment variables
func Process(prefix string, spec interface{}) error {
	return Visit(prefix, spec, func(field reflect.Value, tof reflect.StructField, fieldName string, key string, alt_key string, def string, req string) error {
		value, ok := syscall.Getenv(key)
		if !ok && alt_key != "" {
			key := strings.ToUpper(alt_key)
			value, ok = syscall.Getenv(key)
		}

		if def != "" && !ok {
			value = def
		}

		if !ok && def == "" {
			if req == TrueString {
				return fmt.Errorf("required key %s missing value", key)
			}
			return nil
		}

		err := processField(value, field)
		if err != nil {
			return &ParseError{
				KeyName:   key,
				FieldName: fieldName,
				TypeName:  field.Type().String(),
				Value:     value,
			}
		}

		return nil
	})
}

// MustProcess is the same as Process but panics if an error occurs
func MustProcess(prefix string, spec interface{}) {
	if err := Process(prefix, spec); err != nil {
		panic(err)
	}
}

func processField(value string, field reflect.Value) error {
	typ := field.Type()

	decoder := decoderFrom(field)
	if decoder != nil {
		return decoder.Decode(value)
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if field.IsNil() {
			field.Set(reflect.New(typ))
		}
		field = field.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)
		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, typ.Bits())
		}
		if err != nil {
			return err
		}

		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Slice:
		vals := strings.Split(value, ",")
		sl := reflect.MakeSlice(typ, len(vals), len(vals))
		for i, val := range vals {
			err := processField(val, sl.Index(i))
			if err != nil {
				return err
			}
		}
		field.Set(sl)
	}

	return nil
}

func decoderFrom(field reflect.Value) Decoder {
	if field.CanInterface() {
		dec, ok := field.Interface().(Decoder)
		if ok {
			return dec
		}
	}

	// also check if pointer-to-type implements Decoder,
	// and we can get a pointer to our field
	if field.CanAddr() {
		field = field.Addr()
		dec, ok := field.Interface().(Decoder)
		if ok {
			return dec
		}
	}

	return nil
}
