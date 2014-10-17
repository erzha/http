package server

import (
	"testing"

	"github.com/erzha/kernel"
)

func testStartServer(t *testing.T) {
	h := &Handler{}
	kernel.Boot(h)
}
