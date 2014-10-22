// Copyright 2014 The Wgf Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package session

import (
	"testing"
)

//use default session handler
func TestSet(t *testing.T) {
	sessionHandler = newDefaultHandler()
	ret := &Session{hasStarted: false}

	var ok bool

	//set
	ok = ret.Set("name", "wgf")
	if !ok {
		t.Error("set error")
	}

	//get
	var val string
	ok = ret.Get("name", &val)

	if !ok {
		t.Error("content missed after set")
	}
	if val != "wgf" {
		t.Error("content mismatched after set")
	}
}
