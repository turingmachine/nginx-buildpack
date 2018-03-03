package integration_test

import (
	"nginx/int2/cflocal"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testObject5(t *testing.T, when spec.G, it spec.S) {
	var app *cflocal.App
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
				cutlass.DeleteBuildpack(buildpackName)
			}
		})

		it.Before(func() {
			// TODO uncomment below
			// if !ApiHasMultiBuildpack() {
			// 	t.Skip("Multi buildpack support is required")
			// }

			buildpackName = "override_yml_" + cutlass.RandStringRunes(5)
			g.Expect(cutlass.CreateOrUpdateBuildpack(buildpackName, filepath.Join(bpDir, "fixtures", "overrideyml_bp"))).To(Succeed())

			app, err = cflocal.NewApp(bpDir, "mainline")
			app.Buildpacks = []string{buildpackName + "_buildpack", "nginx_buildpack"}
		})

		it("Forces nginx from override buildpack", func() {
			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.Stdout.String()).To(ContainSubstring("-----> OverrideYML Buildpack"))
			g.Expect(app.ConfirmBuildpack(packagedBuildpack.Version)).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing nginx"))
			g.Eventually(app.Stdout.String).Should(MatchRegexp("Copy .*/nginx.tgz"))
			g.Eventually(app.Stdout.String).Should(ContainSubstring("Could not install nginx: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc, actual sha256 b56b58ac21f9f42d032e1e4b8bf8b8823e69af5411caa15aee2b140bc756962f"))
		})
	})
}
