package integration_test

import (
	"nginx/int2/cfapi"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestApps(t *testing.T) {
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

		when("with no specified version", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "unspecified_version")
				Expect(err).ToNot(HaveOccurred())
			})

			it("Uses latest mainline nginx", func() {
				Expect(app.PushAndConfirm()).To(Succeed())

				Eventually(app.Log).Should(ContainSubstring(`No nginx version specified - using mainline => 1.13.`))
				Eventually(app.Log).ShouldNot(ContainSubstring(`Requested nginx version:`))

				Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
				Eventually(app.Log).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
			})
		})

		when("with an nginx app specifying mainline", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "mainline")
				Expect(err).ToNot(HaveOccurred())
			})

			it("Logs nginx buildpack version", func() {
				Expect(app.PushAndConfirm()).To(Succeed())

				Eventually(app.Log).Should(ContainSubstring(`Requested nginx version: mainline => 1.13.`))

				Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
				Eventually(app.Log).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
			})
		})

		when("with an nginx app specifying stable", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "stable")
				Expect(err).ToNot(HaveOccurred())
			})

			it("Logs nginx buildpack version", func() {
				Expect(app.PushAndConfirm()).To(Succeed())

				Eventually(app.Log).Should(ContainSubstring(`Requested nginx version: stable => 1.12.`))
				Eventually(app.Log).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))

				Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
				Eventually(app.Log).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
			})
		})

		when("with an nginx app specifying 1.12.x", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "1_12_x")
				Expect(err).ToNot(HaveOccurred())
			})

			it("Logs nginx buildpack version", func() {
				Expect(app.PushAndConfirm()).To(Succeed())

				Eventually(app.Log).Should(ContainSubstring(`Requested nginx version: 1.12.x => 1.12.`))
				Eventually(app.Log).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))

				Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
				Eventually(app.Log).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
			})
		})

		when("with an nginx app specifying an unknown version", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "unavailable_version")
				Expect(err).ToNot(HaveOccurred())
			})

			it("Logs nginx buildpack version", func() {
				Expect(app.Push()).ToNot(Succeed())

				Eventually(app.Log).Should(ContainSubstring(`Available versions: mainline, stable, 1.12.x, 1.13.x`))
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
