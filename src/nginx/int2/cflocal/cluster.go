package cflocal

import (
	"io/ioutil"
	"path/filepath"
)

type Cluster struct {
	buildpacks              map[string]string
	defaultBuildpackName    string
	defaultBuildpackVersion string
}

func NewCluster() *Cluster {
	return &Cluster{
		buildpacks: map[string]string{},
	}
}

func (c *Cluster) UploadBuildpack(name, version, file string) error {
	c.buildpacks[name] = file
	if len(c.buildpacks) == 1 {
		c.defaultBuildpackName = name
		c.defaultBuildpackVersion = version
	}
	return nil
}

func (c *Cluster) DeleteBuildpack(name string) error {
	delete(c.buildpacks, name)
	return nil
}

func (c *Cluster) NewApp(bpDir, fixtureName string) (*App, error) {
	tmpPath, err := ioutil.TempDir("", "cflocal.app.")
	if err != nil {
		return nil, err
	}
	return &App{
		cluster:    c,
		Buildpacks: []string{},
		fixture:    filepath.Join(bpDir, "fixtures", fixtureName),
		name:       fixtureName,
		tmpPath:    tmpPath,
	}, nil
}

func (c *Cluster) buildpack(buildpack string) string {
	if c.buildpacks[buildpack] != "" {
		return c.buildpacks[buildpack]
	}
	return buildpack
}

func (c *Cluster) HasMultiBuildpack() bool {
	return true
}
