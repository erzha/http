// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"testing"

	"github.com/erzha/kernel"
)

func testStartServer(t *testing.T) {
	h := &Handler{}
	kernel.Boot(h)
}
