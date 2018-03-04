package integration_test

import (
	"fmt"
	"nginx/int2/cfapi"
	"nginx/int2/cfapi/pack"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var bpDir string
var cluster cfapi.Cluster

func Test(t *testing.T) {
	cutlass.SeedRandom()

	var err error
	bpDir, err = cutlass.FindRoot()
	if err != nil {
		t.Error(fmt.Errorf("Could not find buildpack root dir: %s", err))
	}

	// TODO allow choosing which cluster to use
	// cluster = cflocal.NewCluster()
	cluster = pack.NewCluster()

	// buildpack, err := cutlass.PackageUniquelyVersionedBuildpack()
	// if err != nil {
	// 	t.Error(fmt.Errorf("Could not build buildpack: %s", err))
	// }
	// if err := cluster.UploadBuildpack("nginx_buildpack", buildpack.Version, buildpack.File); err != nil {
	// 	t.Error(fmt.Errorf("Could not upload default buildpack: %s", err))
	// }

	// TODO use the above instead
	if err := cluster.UploadBuildpack("nginx_buildpack", "0.0.4", filepath.Join(bpDir, "nginx_buildpack-cached-v0.0.4.zip")); err != nil {
		panic(fmt.Errorf("Could not upload default buildpack: %s", err))
	}

	spec.Run(t, "Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		testObject3(t, when, it)
		testObject4(t, when, it)
		testObject5(t, when, it)
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
