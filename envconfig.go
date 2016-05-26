// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"errors"
	"fmt"
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
	s := reflect.ValueOf(spec)

	if s.Kind() != reflect.Ptr {
		return ErrInvalidSpecification
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return ErrInvalidSpecification
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fSpec := s.Type().Field(i)
		if !f.CanSet() || fSpec.Tag.Get("ignored") == "true" {
			continue
		}

		if fSpec.Anonymous && f.Kind() == reflect.Struct {
			embeddedPtr := f.Addr().Interface()
			if err := Process(prefix, embeddedPtr); err != nil {
				return err
			}
			f.Set(reflect.ValueOf(embeddedPtr).Elem())
		}

		if err := processField(f, fSpec, prefix); err != nil {
			return err
		}
	}
	return nil
}

// MustProcess is the same as Process but panics if an error occurs
func MustProcess(prefix string, spec interface{}) {
	if err := Process(prefix, spec); err != nil {
		panic(err)
	}
}

// processField processes a field using the given spec/prefix.
func processField(field reflect.Value, spec reflect.StructField, prefix string) error {
	alt := spec.Tag.Get("envconfig")
	fieldName := spec.Name
	if alt != "" {
		fieldName = alt
	}

	var key string
	if spec.Anonymous && alt == "" {
		key = strings.ToUpper(prefix)
	} else {
		key = strings.ToUpper(fmt.Sprintf("%s_%s", prefix, fieldName))
	}

	switch field.Kind() {
	case reflect.Struct:
		return Process(key, field.Addr().Interface())
	case reflect.Ptr:
		fieldType := spec.Type.Elem()
		if fieldType.Kind() != reflect.Struct {
			break
		}

		if field.IsNil() {
			fieldValue := reflect.New(fieldType)
			err := Process(key, fieldValue.Interface())

			if fieldValue.IsValid() {
				field.Set(fieldValue)
			}

			return err
		}

		return Process(key, field.Interface())
	}

	// `os.Getenv` cannot differentiate between an explicitly set empty value
	// and an unset value. `os.LookupEnv` is preferred to `syscall.Getenv`,
	// but it is only available in go1.5 or newer.
	value, ok := syscall.Getenv(key)
	if !ok && alt != "" {
		key := strings.ToUpper(fieldName)
		value, ok = syscall.Getenv(key)
	}

	def := spec.Tag.Get("default")
	if def != "" && !ok {
		value = def
	}

	req := spec.Tag.Get("required")
	if !ok && def == "" {
		if req == "true" {
			return fmt.Errorf("required key %s missing value", key)
		}
		return nil
	}

	if err := processFieldValue(field, value); err != nil {
		return &ParseError{
			KeyName:   key,
			FieldName: fieldName,
			TypeName:  field.Type().String(),
			Value:     value,
		}
	}
	return nil
}

// processFieldValue processes a field using the given value.
func processFieldValue(field reflect.Value, value string) error {
	if decoder := decoderFrom(field); decoder != nil {
		return decoder.Decode(value)
	}

	typ := field.Type()

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
			err := processFieldValue(sl.Index(i), val)
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
