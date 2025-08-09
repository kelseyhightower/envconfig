//go:build go1.18
// +build go1.18

package envconfig

import (
	"os"
	"testing"
	"time"
)

// A GenericSetter has a Set method accepting an already-decoded value of type T.
//
// The GenericSetter interface is not defined outside of this go1.18-gated test file because:
// - Generics are not available until go1.18, and this library currently supports much older code.
// - It is not useful in Process - we cannot type-assert GenericSetter[T] without knowing T at compile time.
type GenericSetter[T any] interface {
	// Set accepts an already-decoded value T.
	//
	// The value should handle its own decoding via any envconfig-supported mechanism.
	// Because value should be fully decoded, there is no error return on this method, unlike the flag-compatible
	// Set method accepting a string.
	Set(value T)
}

type Box[T any] struct {
	Value T
}

func (b *Box[T]) Set(value T) {
	b.Value = value
}

// A Box is an example of a GenericSetter.
var _ GenericSetter[int] = (*Box[int])(nil)

type StringBox struct {
	Value string
}

func (b *StringBox) Set(value string) {
	b.Value = value
}

// Though not generic, StringBox implements GenericSetter[string].
//
// We support non-generic wrapper types with a Set method.
var _ GenericSetter[string] = (*StringBox)(nil)

type SpecWithGenericSetters struct {
	IntBox       Box[int]
	IntBoxPtr    *Box[int]
	IntPtrBox    Box[*int]
	IntPtrBoxPtr *Box[*int]
	MapBox       Box[map[string]time.Duration]
	StringBox    StringBox
}

func TestProcess_GenericSetter(t *testing.T) {
	var s SpecWithGenericSetters
	os.Clearenv()
	os.Setenv("ENV_CONFIG_INTBOX", "1")
	os.Setenv("ENV_CONFIG_INTBOXPTR", "2")
	os.Setenv("ENV_CONFIG_INTPTRBOX", "3")
	os.Setenv("ENV_CONFIG_INTPTRBOXPTR", "4")
	os.Setenv("ENV_CONFIG_MAPBOX", "short:1s,medium:2s,long:3s")
	os.Setenv("ENV_CONFIG_STRINGBOX", "mystring")
	err := Process("env_config", &s)
	if err != nil {
		t.Error(err.Error())
	}
	if s.IntBox.Value != 1 {
		t.Errorf("expected %v, got %v", "1", s.IntBox.Value)
	}
	if s.IntBoxPtr.Value != 2 {
		t.Errorf("expected %v, got %v", "2", s.IntBoxPtr.Value)
	}
	if *s.IntPtrBox.Value != 3 {
		t.Errorf("expected %v, got %v", "3", s.IntPtrBox.Value)
	}
	if *s.IntPtrBoxPtr.Value != 4 {
		t.Errorf("expected %v, got %v", "4", s.IntPtrBoxPtr.Value)
	}
	if len(s.MapBox.Value) != 3 ||
		s.MapBox.Value["short"] != 1*time.Second ||
		s.MapBox.Value["medium"] != 2*time.Second ||
		s.MapBox.Value["long"] != 3*time.Second {
		t.Errorf(
			"expected %#v, got %#v",
			map[string]time.Duration{
				"short":  1 * time.Second,
				"medium": 2 * time.Second,
				"long":   3 * time.Second,
			},
			s.MapBox.Value,
		)
	}
	if s.StringBox.Value != "mystring" {
		t.Errorf("expected %v, got %v", "mystring", s.StringBox.Value)
	}
}
