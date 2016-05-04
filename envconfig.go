// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ErrInvalidSpecification indicates that a specification is of the wrong type.
var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

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

func (e *ParseError) Error() string {
	return fmt.Sprintf("envconfig.Process: assigning %[1]s to %[2]s: converting '%[3]s' to type %[4]s", e.KeyName, e.FieldName, e.Value, e.TypeName)
}

// Process populates the specified struct based on environment variables
func Process(prefix string, spec interface{}) error {
	return ProcessWithUsage(prefix, spec, false)
}

// Process populates the specified struct based on environment variables or displays
// a usage message
func ProcessWithUsage(prefix string, spec interface{}, usage bool) error {
	s := reflect.ValueOf(spec)

	if s.Kind() != reflect.Ptr {
		return ErrInvalidSpecification
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return ErrInvalidSpecification
	}
	typeOfSpec := s.Type()

	if usage {
		fmt.Printf("USAGE: %s\n", path.Base(os.Args[0]))
		fmt.Println()
		fmt.Println("  This application is configured via the environment. The following environment")
		fmt.Println("  variables can used specified:")
		fmt.Println()
	}

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if !f.CanSet() || typeOfSpec.Field(i).Tag.Get("ignored") == "true" {
			continue
		}

		if typeOfSpec.Field(i).Anonymous && f.Kind() == reflect.Struct {
			embeddedPtr := f.Addr().Interface()
			if err := ProcessWithUsage(prefix, embeddedPtr, usage); err != nil {
				return err
			}
			f.Set(reflect.ValueOf(embeddedPtr).Elem())
		}

		alt := typeOfSpec.Field(i).Tag.Get("envconfig")
		fieldName := typeOfSpec.Field(i).Name
		if alt != "" {
			fieldName = alt
		}

		var key string
		if prefix == "" {
			key = strings.ToUpper(fieldName)
		} else {
			key = strings.ToUpper(fmt.Sprintf("%s_%s", prefix, fieldName))
		}

		if usage {
			fmt.Printf("  %s\n", key)
			fmt.Printf("    [description] %s\n", typeOfSpec.Field(i).Tag.Get("short"))
			fmt.Printf("    [type]        %s\n", typeOfSpec.Field(i).Type.Name())
			fmt.Printf("    [default]     %s\n", typeOfSpec.Field(i).Tag.Get("default"))
			isRequired := typeOfSpec.Field(i).Tag.Get("required")
			if isRequired == "" {
				isRequired = "false"
			}
			fmt.Printf("    [required]    %s\n", isRequired)
			continue
		}

		// `os.Getenv` cannot differentiate between an explicitly set empty value
		// and an unset value. `os.LookupEnv` is preferred to `syscall.Getenv`,
		// but it is only available in go1.5 or newer.
		value, ok := syscall.Getenv(key)
		if !ok && alt != "" {
			key := strings.ToUpper(fieldName)
			value, ok = syscall.Getenv(key)
		}

		def := typeOfSpec.Field(i).Tag.Get("default")
		if def != "" && !ok {
			value = def
		}

		req := typeOfSpec.Field(i).Tag.Get("required")
		if !ok && def == "" {
			if req == "true" {
				return fmt.Errorf("required key %s missing value", key)
			}
			continue
		}

		err := processField(value, f)
		if err != nil {
			return &ParseError{
				KeyName:   key,
				FieldName: fieldName,
				TypeName:  f.Type().String(),
				Value:     value,
			}
		}
	}
	return nil
}

// MustProcessWithUsage is the same as Process but panics if an error occurs
func MustProcessWithUsage(prefix string, spec interface{}, usage bool) {
	if err := ProcessWithUsage(prefix, spec, usage); err != nil {
		panic(err)
	}
}

// MustProcess is the same as Process but panics if an error occurs
func MustProcess(prefix string, spec interface{}) {
	MustProcessWithUsage(prefix, spec, false)
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
