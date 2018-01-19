package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("CF Nginx Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("with a simple nginx app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple"))
		})

		It("Uses latested mainline nginx", func() {
			PushAppAndConfirm(app)

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version:  => 1.13.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	Context("with an nginx app specifying mainline", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "mainline"))
		})

		FIt("Logs nginx buildpack version", func() {
			PushAppAndConfirm(app)

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: mainline => 1.13.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	Context("with an nginx app specifying stable", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "stable"))
		})

		It("Logs nginx buildpack version", func() {
			PushAppAndConfirm(app)

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: stable => 1.12.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	Context("with an nginx app specifying 1.12.x", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "1_12_x"))
		})

		It("Logs nginx buildpack version", func() {
			PushAppAndConfirm(app)

			Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: 1.12.x => 1.12.`))

			Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	Context("with an nginx app specifying 1.1.x", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "unknown_version"))
		})

		It("Logs nginx buildpack version", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring(`Available versions: mainline, stable, 1.13.x, 1.12.x,`))
		})
	})
})
