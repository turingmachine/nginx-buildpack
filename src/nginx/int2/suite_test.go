package integration_test

import (
	"fmt"
	"nginx/int2/cfapi"
	"nginx/int2/cfapi/cflocal"
	"nginx/int2/cfapi/foundation"
	"nginx/int2/cfapi/pack"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var bpDir string
var cluster cfapi.Cluster

func Test(t *testing.T) {
	var err error
	bpDir, err = cutlass.FindRoot()
	if err != nil {
		t.Error(fmt.Errorf("Could not find buildpack root dir: %s", err))
	}

	if err := cutlass.CopyCfHome(); err != nil {
		t.Error(fmt.Errorf("Could not copy cf home dir: %s", err))
	}
	cutlass.SeedRandom()
	gomega.SetDefaultEventuallyTimeout(10 * time.Second)

	// TODO allow choosing which cluster to use
	if true {
		cluster = pack.NewCluster()
	} else if false {
		cluster = cflocal.NewCluster()
	} else {
		cluster = foundation.NewCluster()
	}

	// cutlass.Cached = true
	// fmt.Println("Building Buildpack")
	// buildpack, err := cutlass.PackageUniquelyVersionedBuildpack()
	// if err != nil {
	// 	t.Error(fmt.Errorf("Could not build buildpack: %s", err))
	// }
	// fmt.Println("Uploading Buildpack:", buildpack.File)
	// if err := cluster.UploadBuildpack("nginx_buildpack", buildpack.Version, buildpack.File); err != nil {
	// 	t.Error(fmt.Errorf("Could not upload default buildpack: %s", err))
	// }

	// TODO use the above instead
	fmt.Println("Uploading Buildpack")
	if err := cluster.UploadBuildpack("nginx_buildpack", "0.0.4.20180305091047", filepath.Join(bpDir, "nginx_buildpack-cached-v0.0.4.20180305091047.zip")); err != nil {
		panic(fmt.Errorf("Could not upload default buildpack: %s", err))
	}

	spec.Run(t, "Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		testObject3(t, when, it)
		testObject4(t, when, it)
		testObject5(t, when, it)
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
