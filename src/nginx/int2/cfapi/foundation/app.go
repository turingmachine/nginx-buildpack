package foundation

import (
	"bytes"
	"nginx/int2/cfapi/utils"
)

type App struct {
	cluster    *Cluster
	buildpacks []string
	fixture    string
	name       string
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
}

func (a *App) Buildpacks(buildpacks []string) {
	a.buildpacks = buildpacks
}

func (a *App) ConfirmBuildpack(version string) error {
	// TODO
	return nil
}

func (a *App) Push() error {
	return nil
}

func (a *App) PushAndConfirm() error {
	if err := a.Push(); err != nil {
		return err
	}
	// TODO
	// Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
	return nil
}

func (a *App) Stop() error {
	return nil
}

func (a *App) Destroy() error {
	return nil
}

func (a *App) GetUrl(path string) (string, error) {
	// TODO
	return "", nil
}

func (a *App) Get(path string, headers map[string]string) (string, map[string][]string, error) {
	url, err := a.GetUrl(path)
	if err != nil {
		return "", map[string][]string{}, err
	}
	return utils.HttpGet(url, headers)
}

func (a *App) GetBody(path string) (string, error) {
	url, err := a.GetUrl(path)
	if err != nil {
		return "", err
	}
	return utils.HttpGetBody(url)
}

func (a *App) Log() string {
	return a.Stdout.String()
}
