// Copyright (c) 2017 Jerry Jacobs. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"io"
	"io/ioutil"
)

const (
	decoderStateKey = iota
	decoderStateValue
)

// NewReader creates a new envconfig process based on the contents of r
func NewReader(r io.Reader) (ProcessFunc, error) {
	le, err := newReaderLookupEnvFunc(r)
	if err != nil {
		return nil, err
	}
	return func(prefix string, spec interface{}) error {
		return process(le, prefix, spec)
	}, nil
}

// newReaderLookupEnvFunc creates a new LookupEnvFunc based on contents of r
func newReaderLookupEnvFunc(r io.Reader) (LookupEnvFunc, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	environ := make(map[string]string)

	var state int
	var buf []byte
	var key string
	var val string

	// collect
	for _, b := range data {
		switch b {
		case '_':
			if state == decoderStateValue {
				buf = append(buf, b)
			}
		case '=':
			key = string(buf)
			state = decoderStateValue
			buf = nil
		case '\n':
			val = string(buf)
			buf = nil
			environ[key] = val
			state = decoderStateKey
		default:
			buf = append(buf, b)
		}
	}

	// collect last buffered item without a newline
	if len(buf) > 0 && state == decoderStateValue {
		environ[key] = string(buf)
	}

	return func(key string) (string, bool) {
		value, ok := environ[key]
		return value, ok
	}, nil
}
