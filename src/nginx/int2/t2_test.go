package integration_test

import (
	"testing"
	"time"

	"github.com/sclevine/spec"
)

func testObject2(t *testing.T, when spec.G, it spec.S) {
	var pause time.Duration

	it.Before(func() {
		pause = 1 * time.Second
	})

	it("do thing", func() {
		t.Error("bad default")
	})

	it("thing 1", func() {
		time.Sleep(pause)
	})
	it("thing 2", func() {
		time.Sleep(pause)
	})
	it("thing 3", func() {
		time.Sleep(pause)
	})
}
