package integration_test

import (
	"fmt"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var bpDir string
var packagedBuildpack cutlass.VersionedBuildpackPackage

func Test(t *testing.T) {
	var err error
	bpDir, err = cutlass.FindRoot()
	if err != nil {
		t.Error(fmt.Errorf("Could not find buildpack root dir: %s", err))
	}

	packagedBuildpack, err = cutlass.PackageUniquelyVersionedBuildpack()
	if err != nil {
		t.Error(fmt.Errorf("Could not build buildpack: %s", err))
	}

	spec.Run(t, "Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		testObject3(t, when, it)
		testObject4(t, when, it)
		testObject5(t, when, it)
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
