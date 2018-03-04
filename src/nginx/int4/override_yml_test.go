package integration_test

import (
	"nginx/int2/cflocal"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/gomega"
)

func Test3(t *testing.T) {
	t.Parallel()
	When(t, "override yml", func(t *testing.T) {
		if !cluster.HasMultiBuildpack() {
			t.Skip("Multi buildpack support is required")
		}
		SimpleTest(t, "Forces nginx from override buildpack", "mainline", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			buildpackName := "override_yml_" + cutlass.RandStringRunes(5)
			g.Expect(cluster.UploadBuildpack(buildpackName, "", filepath.Join(bpDir, "fixtures", "overrideyml_bp"))).To(Succeed())
			defer cluster.DeleteBuildpack(buildpackName)
			app.Buildpacks = []string{buildpackName, "nginx_buildpack"}

			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.Stdout.String()).To(ContainSubstring("-----> OverrideYML Buildpack"))
			g.Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing nginx"))
			g.Eventually(app.Stdout.String).Should(MatchRegexp("Copy .*/nginx.tgz"))
			g.Eventually(app.Stdout.String).Should(ContainSubstring("Could not install nginx: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc, actual sha256 b56b58ac21f9f42d032e1e4b8bf8b8823e69af5411caa15aee2b140bc756962f"))
		})
	})
}
