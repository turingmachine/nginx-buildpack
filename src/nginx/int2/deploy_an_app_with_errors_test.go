package integration_test

import (
	"nginx/int2/cfapi"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testObject4(t *testing.T, when spec.G, it spec.S) {
	var app cfapi.App
	var err error
	var g *GomegaWithT
	it.Before(func() { g = NewGomegaWithT(t) })

	it.After(func() {
		if app != nil {
			app.Destroy()
		}
	})

	when("an app without nginx.conf", func() {
		it.Before(func() {
			app, err = cluster.NewApp(bpDir, "empty")
			g.Expect(err).ToNot(HaveOccurred())
			app.Buildpacks([]string{"nginx_buildpack"})
		})

		it("Logs nginx an error", func() {
			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

			g.Eventually(app.Log).Should(ContainSubstring("nginx.conf file must be present at the app root"))
		})
	})

	when("an app with nginx.conf without {{.Port}}", func() {
		it.Before(func() {
			app, err = cluster.NewApp(bpDir, "missing_template_port")
			g.Expect(err).ToNot(HaveOccurred())
		})

		it("Logs nginx an error", func() {
			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

			g.Eventually(app.Log).Should(ContainSubstring("nginx.conf file must be configured to respect the value of `{{.Port}}`"))
		})
	})
}
