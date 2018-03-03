package integration_test

import (
	"nginx/int2/cflocal"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testObject3(t *testing.T, when spec.G, it spec.S) {
	var app *cflocal.App
	var err error

	it.After(func() { app.Destroy() })

	when("with no specified version", func() {
		it.Before(func() {
			app, err = cflocal.NewApp(bpDir, "unspecified_version")
			Expect(err).ToNot(HaveOccurred())
		})

		it("Uses latest mainline nginx", func() {
			Expect(app.PushAndConfirm()).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring(`No nginx version specified - using mainline => 1.13.`))
			Eventually(app.Stdout.String).ShouldNot(ContainSubstring(`Requested nginx version:`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	when("with an nginx app specifying mainline", func() {
		it.Before(func() {
			app, err = cflocal.NewApp(bpDir, "mainline")
			Expect(err).ToNot(HaveOccurred())
		})

		it("Logs nginx buildpack version", func() {
			Expect(app.PushAndConfirm()).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: mainline => 1.13.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	when("with an nginx app specifying stable", func() {
		it.Before(func() {
			app, err = cflocal.NewApp(bpDir, "stable")
			Expect(err).ToNot(HaveOccurred())
		})

		it("Logs nginx buildpack version", func() {
			Expect(app.PushAndConfirm()).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: stable => 1.12.`))
			Eventually(app.Stdout.String).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	when("with an nginx app specifying 1.12.x", func() {
		it.Before(func() {
			app, err = cflocal.NewApp(bpDir, "1_12_x")
			Expect(err).ToNot(HaveOccurred())
		})

		it("Logs nginx buildpack version", func() {
			Expect(app.PushAndConfirm()).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: 1.12.x => 1.12.`))
			Eventually(app.Stdout.String).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	when("with an nginx app specifying an unknown version", func() {
		it.Before(func() {
			app, err = cflocal.NewApp(bpDir, "unavailable_version")
			Expect(err).ToNot(HaveOccurred())
		})

		it("Logs nginx buildpack version", func() {
			Expect(app.Push()).ToNot(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring(`Available versions: mainline, stable, 1.12.x, 1.13.x`))
		})
	})
}
