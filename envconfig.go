// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package envconfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ParseError struct {
	FieldName string
	TypeName  string
	Value     string
}

type InvalidSpecificationError struct{}

func (e InvalidSpecificationError) Error() string {
	return "envconfig.Process: invalid specification type must be a struct"
}

func (e ParseError) Error() string {
	return fmt.Sprintf("envconfig.Process: assigning to %s: converting '%s' to an %s", e.FieldName, e.Value, e.TypeName)
}

func Process(prefix string, spec interface{}) error {
	s := reflect.ValueOf(spec).Elem()
	if s.Kind() != reflect.Struct {
		return &InvalidSpecificationError{}
	}
	typeOfSpec := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.CanSet() {
			fieldName := typeOfSpec.Field(i).Name
			key := fmt.Sprintf("%s_%s", prefix, fieldName)
			value := os.Getenv(strings.ToUpper(key))
			switch f.Kind() {
			case reflect.String:
				f.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intValue, err := strconv.ParseInt(value, 0, f.Type().Bits())
				if err != nil {
					return &ParseError{
						FieldName: fieldName,
						TypeName:  f.Kind().String(),
						Value:     value,
					}
				}
				f.SetInt(intValue)
			case reflect.Bool:
				boolValue, err := strconv.ParseBool(value)
				if err != nil {
					return &ParseError{
						FieldName: fieldName,
						TypeName:  f.Kind().String(),
						Value:     value,
					}
				}
				f.SetBool(boolValue)
			case reflect.Float32:
				floatValue, err := strconv.ParseFloat(value, f.Type().Bits())
				if err != nil {
					return err
				}
				f.SetFloat(floatValue)
			}
		}
	}
	return nil
}
