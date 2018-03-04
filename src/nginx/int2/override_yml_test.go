package integration_test

import (
	"nginx/int2/cfapi"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testObject5(t *testing.T, when spec.G, it spec.S) {
	var app cfapi.App
	var err error
	var g *GomegaWithT
	it.Before(func() { g = NewGomegaWithT(t) })

	it.After(func() {
		if app != nil {
			app.Destroy()
		}
	})

	when("override yml", func() {
		var buildpackName string
		it.After(func() {
			if buildpackName != "" {
				cluster.DeleteBuildpack(buildpackName)
			}
		})

		it.Before(func() {
			if !cluster.HasMultiBuildpack() {
				t.Skip("Multi buildpack support is required")
			}

			buildpackName = "override_yml_" + cutlass.RandStringRunes(5)
			g.Expect(cluster.UploadBuildpack(buildpackName, "", filepath.Join(bpDir, "fixtures", "overrideyml_bp"))).To(Succeed())

			app, err = cluster.NewApp(bpDir, "mainline")
			app.Buildpacks([]string{buildpackName, "nginx_buildpack"})
		})

		it("Forces nginx from override buildpack", func() {
			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.Log()).To(ContainSubstring("-----> OverrideYML Buildpack"))
			g.Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

			g.Eventually(app.Log).Should(ContainSubstring("-----> Installing nginx"))
			g.Eventually(app.Log).Should(MatchRegexp("Copy .*/nginx.tgz"))
			g.Eventually(app.Log).Should(ContainSubstring("Could not install nginx: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc, actual sha256 b56b58ac21f9f42d032e1e4b8bf8b8823e69af5411caa15aee2b140bc756962f"))
		})
	})
}
