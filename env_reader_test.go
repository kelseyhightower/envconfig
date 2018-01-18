// Copyright (c) 2017 Jerry Jacobs. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReader(t *testing.T) {
	const tv = `USER=xor-gate`

	envLookup, err := newReaderLookupEnvFunc(strings.NewReader(tv))
	assert.Nil(t, err)

	user, ok := envLookup("USER")
	assert.True(t, ok)
	assert.Equal(t, "xor-gate", user)
}
