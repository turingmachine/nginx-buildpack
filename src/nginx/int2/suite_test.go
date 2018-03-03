package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var bpDir string

func Test(t *testing.T) {
	var err error
	bpDir, err = cutlass.FindRoot()
	if err != nil {
		t.Error("Could not find buildpack root dir")
	}

	spec.Run(t, "Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() { RegisterTestingT(t) })

		testObject3(t, when, it)
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
