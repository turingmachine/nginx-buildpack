package integration_test

import (
	"fmt"
	"nginx/int2/cflocal"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/onsi/gomega"
)

var bpDir string
var cluster *cflocal.Cluster

func init() {
	var err error
	bpDir, err = cutlass.FindRoot()
	if err != nil {
		panic(fmt.Errorf("Could not find buildpack root dir: %s", err))
	}
	cluster = cflocal.NewCluster()

	// buildpack, err := cutlass.PackageUniquelyVersionedBuildpack()
	// if err != nil {
	// 	panic(fmt.Errorf("Could not build buildpack: %s", err))
	// }
	// if err := cluster.UploadBuildpack("nginx_buildpack", buildpack.Version, buildpack.File); err != nil {
	// 	panic(fmt.Errorf("Could not upload default buildpack: %s", err))
	// }

	if err := cluster.UploadBuildpack("nginx_buildpack", "0.0.4", "/Users/dgodd/workspace/nginx-buildpack/nginx_buildpack-cached-v0.0.4.zip"); err != nil {
		panic(fmt.Errorf("Could not upload default buildpack: %s", err))
	}

	gomega.SetDefaultEventuallyTimeout(10 * time.Second)
}

func When(t *testing.T, name string, f func(*testing.T)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		f(t)
	})
}

func SimpleTest(t *testing.T, name string, fixture string, f func(*testing.T, *gomega.GomegaWithT, *cflocal.App)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		app, err := cluster.NewApp(bpDir, fixture)
		if err != nil {
			t.Error(fmt.Printf("Could not create app %s: %s", fixture, err))
			return
		}
		defer app.Destroy()

		g := gomega.NewGomegaWithT(t)
		f(t, g, app)
	})
}
