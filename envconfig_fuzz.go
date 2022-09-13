// +build gofuzz

package envconfig

import (
	"fmt"
	"os"
	"time"
)

const (
	prefix = "FUZZIT"
)

type Specification struct {
	String        string
	Int8          int8
	Int16         int16
	Int32         int32
	Int64         int64
	Bool          bool
	Float32       float32
	Float64       float64
	Duration      time.Duration
	SliceString   []string
	SliceInt8     []int8
	SliceInt16    []int16
	SliceInt32    []int32
	SliceInt64    []int64
	SliceBool     []bool
	SliceFloat32  []float32
	SliceFloat64  []float64
	SliceDuration []time.Duration
	Map           map[string]string
	Embedded      struct {
		String      string
		Int8        int8
		Bool        bool
		Float32     float32
		Duration    time.Duration
		SliceString []string
		Map         map[string]string
	}
}

func Fuzz(fuzz []byte) int {
	ok := stage(fuzz)
	if !ok {
		return -1
	}
	specification := Specification{}
	Process(prefix, specification)
	return 0
}

func stage(fuzz []byte) (ok bool) {
	if len(fuzz) < 3 {
		return false
	}
	var value string
	for len(fuzz) != 0 {
		if fuzz[0] > 19 {
			return false
		}
		name := selectName(fuzz[0])
		value, fuzz, ok = extractValue(fuzz)
		if !ok {
			return false
		}
		err := os.Setenv(fmt.Sprintf("%s_%s", prefix, name), value)
		if err != nil {
			return false
		}
	}
	return true
}

func extractValue(fuzz []byte) (value string, rest []byte, ok bool) {
	if len(fuzz) < 3 {
		return "", nil, false
	}
	length := int(fuzz[1])
	if length == 0 {
		return "", nil, false
	}
	if len(fuzz) < (length + 2) {
		return "", nil, false
	}
	value = string(fuzz[2 : length+2])
	rest = fuzz[length+2:]
	return value, rest, true
}

func selectName(selector byte) (name string) {
	switch selector {
	case 0:
		return "String"
	case 1:
		return "Int8"
	case 2:
		return "Int16"
	case 3:
		return "Int32"
	case 4:
		return "Int64"
	case 5:
		return "Bool"
	case 6:
		return "Float32"
	case 7:
		return "Float64"
	case 8:
		return "Duration"
	case 9:
		return "SliceString"
	case 10:
		return "SliceInt8"
	case 11:
		return "SliceInt16"
	case 12:
		return "SliceInt32"
	case 13:
		return "SliceInt64"
	case 14:
		return "SliceBool"
	case 15:
		return "SliceFloat32"
	case 16:
		return "SliceFloat64"
	case 17:
		return "SliceDuration"
	case 18:
		return "Map"
	case 19:
		return "Embedded"
	default:
		panic("invalid selector")
	}
}
