package integration_test

import (
	"nginx/int2/cflocal"
	"testing"

	. "github.com/onsi/gomega"
)

func Test1(t *testing.T) {
	t.Parallel()
	When(t, "with no specified version", func(t *testing.T) {
		SimpleTest(t, "Uses latest mainline nginx", "unspecified_version", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			g.Expect(app.PushAndConfirm()).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring(`No nginx version specified - using mainline => 1.13.`))
			g.Eventually(app.Stdout.String).ShouldNot(ContainSubstring(`Requested nginx version:`))

			g.Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			g.Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	When(t, "with an nginx app specifying mainline", func(t *testing.T) {
		SimpleTest(t, "Logs nginx buildpack version", "mainline", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			g.Expect(app.PushAndConfirm()).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: mainline => 1.13.`))

			g.Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			g.Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	When(t, "with an nginx app specifying stable", func(t *testing.T) {
		SimpleTest(t, "Logs nginx buildpack version", "stable", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			g.Expect(app.PushAndConfirm()).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: stable => 1.12.`))
			g.Eventually(app.Stdout.String).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))

			g.Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			g.Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	When(t, "with an nginx app specifying 1.12.x", func(t *testing.T) {
		SimpleTest(t, "Logs nginx buildpack version", "1_12_x", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			g.Expect(app.PushAndConfirm()).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring(`Requested nginx version: 1.12.x => 1.12.`))
			g.Eventually(app.Stdout.String).Should(ContainSubstring(`Warning: usage of "stable" versions of NGINX is discouraged in most cases by the NGINX team.`))

			g.Expect(app.GetBody("/")).To(ContainSubstring("Exciting Content"))
			g.Eventually(app.Stdout.String).Should(ContainSubstring(`NginxLog "GET / HTTP/1.1" 200`))
		})
	})

	When(t, "with an nginx app specifying an unknown version", func(t *testing.T) {
		SimpleTest(t, "Logs nginx buildpack version", "unavailable_version", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			g.Expect(app.Push()).ToNot(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring(`Available versions: mainline, stable, 1.12.x, 1.13.x`))
		})
	})
}
