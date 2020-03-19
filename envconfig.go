// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidSpecification indicates that a specification is of the wrong type.
var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

var gatherRegexp = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
var acronymRegexp = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")

// A ParseError occurs when an environment variable cannot be converted to
// the type required by a struct field during assignment.
type ParseError struct {
	KeyName   string
	FieldName string
	TypeName  string
	Value     string
	Err       error
}

// Decoder has the same semantics as Setter, but takes higher precedence.
// It is provided for historical compatibility.
type Decoder interface {
	Decode(value string) error
}

// Setter is implemented by types can self-deserialize values.
// Any type that implements flag.Value also implements Setter.
type Setter interface {
	Set(value string) error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("envconfig.Process: assigning %[1]s to %[2]s: converting '%[3]s' to type %[4]s. details: %[5]s", e.KeyName, e.FieldName, e.Value, e.TypeName, e.Err)
}

// varInfo maintains information about the configuration variable
type varInfo struct {
	Name  string
	Alt   string
	Key   string
	Field reflect.Value
	Tags  reflect.StructTag
}

func gatherInfoForUsage(prefix string, spec interface{}) ([]varInfo, error) {
	return gatherInfo(prefix, spec, map[string]string{}, false, true)
}

func gatherInfoForProcessing(prefix string, spec interface{}, env map[string]string) ([]varInfo, error) {
	return gatherInfo(prefix, spec, env, false, false)
}

// gatherInfo gathers information about the specified struct, use gatherInfoForUsage or gatherInfoForProcessing for calling it
func gatherInfo(prefix string, spec interface{}, env map[string]string, isInsideStructSlice, forUsage bool) ([]varInfo, error) {
	s := reflect.ValueOf(spec)

	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidSpecification
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidSpecification
	}
	typeOfSpec := s.Type()

	// over allocate an info array, we will extend if needed later
	infos := make([]varInfo, 0, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)
		if !f.CanSet() || isTrue(ftype.Tag.Get("ignored")) {
			continue
		}

		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					// nil pointer to a non-struct: leave it alone
					break
				}
				// nil pointer to struct: create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		// Capture information about the config variable
		info := varInfo{
			Name:  ftype.Name,
			Field: f,
			Tags:  ftype.Tag,
			Alt:   strings.ToUpper(ftype.Tag.Get("envconfig")),
		}

		// Default to the field name as the env var name (will be upcased)
		info.Key = info.Name

		// Best effort to un-pick camel casing as separate words
		if isTrue(ftype.Tag.Get("split_words")) {
			words := gatherRegexp.FindAllStringSubmatch(ftype.Name, -1)
			if len(words) > 0 {
				var name []string
				for _, words := range words {
					if m := acronymRegexp.FindStringSubmatch(words[0]); len(m) == 3 {
						name = append(name, m[1], m[2])
					} else {
						name = append(name, words[0])
					}
				}

				info.Key = strings.Join(name, "_")
			}
		}
		if info.Alt != "" {
			info.Key = info.Alt
			if isInsideStructSlice {
				// we don't want this to be read, since we're inside of a struct slice,
				// each slice element will have same Alt and thus they would overwrite themselves
				info.Alt = ""
			}
		}
		if prefix != "" {
			info.Key = fmt.Sprintf("%s_%s", prefix, info.Key)
		}
		info.Key = strings.ToUpper(info.Key)

		if decoderFrom(f) != nil || setterFrom(f) != nil || textUnmarshaler(f) != nil || binaryUnmarshaler(f) != nil {
			// there's a decoder defined, no further processing needed
			infos = append(infos, info)
		} else if f.Kind() == reflect.Struct {
			// it's a struct without a specific decoder set
			innerPrefix := prefix
			if !ftype.Anonymous {
				innerPrefix = info.Key
			}

			embeddedPtr := f.Addr().Interface()
			embeddedInfos, err := gatherInfo(innerPrefix, embeddedPtr, env, isInsideStructSlice, forUsage)
			if err != nil {
				return nil, err
			}
			infos = append(infos, embeddedInfos...)
		} else if arePointers := isSliceOfStructPtrs(f); arePointers || isSliceOfStructs(f) {
			// it's a slice of structs
			var (
				l            int
				prefixFormat prefixFormatter
			)
			if forUsage {
				// it's just for usage so we don't know how many of them can be out there
				// so we'll print one info with a generic [N] index
				l = 1
				prefixFormat = usagePrefix{info.Key, "[N]"}
			} else {
				var err error
				// let's find out how many are defined by the env vars, and gather info of each one of them
				if l, err = sliceLen(info.Key, env); err != nil {
					return nil, err
				}
				prefixFormat = processPrefix(info.Key)
				// if no keys, check the alternative keys, unless we're inside of a slice
				if l == 0 && info.Alt != "" && !isInsideStructSlice {
					if l, err = sliceLen(info.Alt, env); err != nil {
						return nil, err
					}
					prefixFormat = processPrefix(info.Alt)
				}
			}

			f.Set(reflect.MakeSlice(f.Type(), l, l))
			for i := 0; i < l; i++ {
				var structPtrValue reflect.Value

				if arePointers {
					f.Index(i).Set(reflect.New(f.Type().Elem().Elem()))
					structPtrValue = f.Index(i)
				} else {
					structPtrValue = f.Index(i).Addr()
				}

				embeddedInfos, err := gatherInfo(prefixFormat.format(i), structPtrValue.Interface(), env, true, forUsage)
				if err != nil {
					return nil, err
				}
				infos = append(infos, embeddedInfos...)
			}
		} else {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

// CheckDisallowed checks that no environment variables with the prefix are set
// that we don't know how or want to parse. This is likely only meaningful with
// a non-empty prefix.
func CheckDisallowed(prefix string, spec interface{}) error {
	env := environment()
	infos, err := gatherInfoForProcessing(prefix, spec, env)
	if err != nil {
		return err
	}

	vars := make(map[string]struct{})
	for _, info := range infos {
		vars[info.Key] = struct{}{}
	}

	if prefix != "" {
		prefix = strings.ToUpper(prefix) + "_"
	}

	for key := range env {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		if _, found := vars[key]; !found {
			return fmt.Errorf("unknown environment variable %s", key)
		}
	}

	return nil
}

// Process populates the specified struct based on environment variables
func Process(prefix string, spec interface{}) error {
	env := environment()
	infos, err := gatherInfoForProcessing(prefix, spec, env)

	for _, info := range infos {
		value, ok := env[info.Key]
		if !ok && info.Alt != "" {
			value, ok = env[info.Alt]
		}

		def := info.Tags.Get("default")
		if def != "" && !ok {
			value = def
		}

		req := info.Tags.Get("required")
		if !ok && def == "" {
			if isTrue(req) {
				key := info.Key
				if info.Alt != "" {
					key = info.Alt
				}
				return fmt.Errorf("required key %s missing value", key)
			}
			continue
		}

		err = processField(value, info.Field)
		if err != nil {
			return &ParseError{
				KeyName:   info.Key,
				FieldName: info.Name,
				TypeName:  info.Field.Type().String(),
				Value:     value,
				Err:       err,
			}
		}
	}

	return err
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
	// look for Set method if Decode not defined
	setter := setterFrom(field)
	if setter != nil {
		return setter.Set(value)
	}

	if t := textUnmarshaler(field); t != nil {
		return t.UnmarshalText([]byte(value))
	}

	if b := binaryUnmarshaler(field); b != nil {
		return b.UnmarshalBinary([]byte(value))
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
		sl := reflect.MakeSlice(typ, 0, 0)
		if typ.Elem().Kind() == reflect.Uint8 {
			sl = reflect.ValueOf([]byte(value))
		} else if len(strings.TrimSpace(value)) != 0 {
			vals := strings.Split(value, ",")
			sl = reflect.MakeSlice(typ, len(vals), len(vals))
			for i, val := range vals {
				err := processField(val, sl.Index(i))
				if err != nil {
					return err
				}
			}
		}
		field.Set(sl)
	case reflect.Map:
		mp := reflect.MakeMap(typ)
		if len(strings.TrimSpace(value)) != 0 {
			pairs := strings.Split(value, ",")
			for _, pair := range pairs {
				kvpair := strings.Split(pair, ":")
				if len(kvpair) != 2 {
					return fmt.Errorf("invalid map item: %q", pair)
				}
				k := reflect.New(typ.Key()).Elem()
				err := processField(kvpair[0], k)
				if err != nil {
					return err
				}
				v := reflect.New(typ.Elem()).Elem()
				err = processField(kvpair[1], v)
				if err != nil {
					return err
				}
				mp.SetMapIndex(k, v)
			}
		}
		field.Set(mp)
	}

	return nil
}

func interfaceFrom(field reflect.Value, fn func(interface{}, *bool)) {
	// it may be impossible for a struct field to fail this check
	if !field.CanInterface() {
		return
	}
	var ok bool
	fn(field.Interface(), &ok)
	if !ok && field.CanAddr() {
		fn(field.Addr().Interface(), &ok)
	}
}

func decoderFrom(field reflect.Value) (d Decoder) {
	interfaceFrom(field, func(v interface{}, ok *bool) { d, *ok = v.(Decoder) })
	return d
}

func setterFrom(field reflect.Value) (s Setter) {
	interfaceFrom(field, func(v interface{}, ok *bool) { s, *ok = v.(Setter) })
	return s
}

func textUnmarshaler(field reflect.Value) (t encoding.TextUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { t, *ok = v.(encoding.TextUnmarshaler) })
	return t
}

func binaryUnmarshaler(field reflect.Value) (b encoding.BinaryUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { b, *ok = v.(encoding.BinaryUnmarshaler) })
	return b
}

func isTrue(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// sliceLen returns the len of a slice of structs defined in the environment config
func sliceLen(prefix string, env map[string]string) (int, error) {
	prefix = prefix + "_"
	indexes := map[int]bool{}
	for k := range env {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		var digits string
		for i := len(prefix); i < len(k); i++ {
			if k[i] >= '0' && k[i] <= '9' {
				digits += k[i : i+1]
			} else if k[i] == '_' {
				break
			} else {
				return 0, fmt.Errorf("key %s has prefix %s but doesn't follow an integer value followed by an underscore (unexpected char %q)", k, prefix, k[i])
			}
		}
		if digits == "" {
			return 0, fmt.Errorf("key %s has prefix %s but doesn't follow an integer value followed by an underscore (no digits found)", k, prefix)
		}
		index, err := strconv.Atoi(digits)
		if err != nil {
			return 0, fmt.Errorf("can't parse index in %s: %s", k, err)
		}
		indexes[index] = true
	}

	for i := 0; i < len(indexes); i++ {
		if _, ok := indexes[i]; !ok {
			return 0, fmt.Errorf("prefix %s defines %d indexes, but index %d is unset: indexes must start at 0 and be consecutive", prefix, len(indexes), i)
		}
	}
	return len(indexes), nil
}

func isSliceOfStructs(v reflect.Value) bool {
	return v.Kind() == reflect.Slice &&
		v.Type().Elem().Kind() == reflect.Struct
}

func isSliceOfStructPtrs(v reflect.Value) bool {
	return v.Kind() == reflect.Slice &&
		v.Type().Elem().Kind() == reflect.Ptr &&
		v.Type().Elem().Elem().Kind() == reflect.Struct
}

func environment() map[string]string {
	environ := os.Environ()
	vars := make(map[string]string, len(environ))
	for _, env := range os.Environ() {
		split := strings.SplitN(env, "=", 2)
		var v string
		if len(split) > 1 {
			v = split[1]
		}
		vars[split[0]] = v
	}
	return vars
}

type prefixFormatter interface{ format(v interface{}) string }

type usagePrefix struct{ prefix, placeholder string }

func (p usagePrefix) format(v interface{}) string { return p.prefix + "_" + p.placeholder }

type processPrefix string

func (p processPrefix) format(v interface{}) string { return fmt.Sprintf(string(p)+"_%d", v) }
