package integration_test

import (
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func Test(t *testing.T) {
	spec.Run(t, "Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		// testObject1(t, when, it)
		// testObject2(t, when, it)
		testObject3(t, when, it)
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testObject1(t *testing.T, when spec.G, it spec.S) {
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
