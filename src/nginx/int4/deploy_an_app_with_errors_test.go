package integration_test

import (
	"nginx/int2/cflocal"
	"testing"

	. "github.com/onsi/gomega"
)

func Test2(t *testing.T) {
	t.Parallel()
	When(t, "an app without nginx.conf", func(t *testing.T) {
		SimpleTest(t, "Logs nginx an error", "empty", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			app.Buildpacks = []string{"nginx_buildpack"}
			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring("nginx.conf file must be present at the app root"))
		})
	})

	When(t, "an app with nginx.conf without {{.Port}}", func(t *testing.T) {
		SimpleTest(t, "Logs nginx an error", "missing_template_port", func(t *testing.T, g *GomegaWithT, app *cflocal.App) {
			g.Expect(app.Push()).ToNot(Succeed())
			g.Expect(app.ConfirmBuildpack("nginx_buildpack")).To(Succeed())

			g.Eventually(app.Stdout.String).Should(ContainSubstring("nginx.conf file must be configured to respect the value of `{{.Port}}`"))
		})
	})
}
