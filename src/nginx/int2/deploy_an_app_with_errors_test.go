package integration_test

import (
	"fmt"
	"nginx/int2/cfapi"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestErrors(t *testing.T) {
	t.Parallel()
	spec.Run(t, "Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		var app cfapi.App
		var err error
		var g *GomegaWithT
		var Expect func(actual interface{}, extra ...interface{}) GomegaAssertion
		var Eventually func(actual interface{}, intervals ...interface{}) GomegaAsyncAssertion
		it.Before(func() {
			g = NewGomegaWithT(t)
			Expect = g.Expect
			Eventually = g.Eventually
		})
		it.After(func() {
			if app != nil {
				app.Destroy()
			}
		})

		when("an app without nginx.conf", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "empty")
				Expect(err).ToNot(HaveOccurred())
				app.Buildpacks([]string{"nginx_buildpack"})
			})

			it("Logs nginx an error", func() {
				Expect(app.Push()).ToNot(Succeed())
				fmt.Println(app.Log())
				Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

				Eventually(app.Log).Should(ContainSubstring("nginx.conf file must be present at the app root"))
			})
		})

		when("an app with nginx.conf without {{.Port}}", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "missing_template_port")
				Expect(err).ToNot(HaveOccurred())
			})

			it("Logs nginx an error", func() {
				Expect(app.Push()).ToNot(Succeed())
				Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

				Eventually(app.Log).Should(ContainSubstring("nginx.conf file must be configured to respect the value of `{{.Port}}`"))
			})
		})
	})
}
