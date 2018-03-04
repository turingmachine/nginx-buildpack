package pack

import (
	"io/ioutil"
	"nginx/int2/cfapi/models"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"
)

type Cluster struct {
	buildpacks              map[string]string
	defaultBuildpackName    string
	defaultBuildpackVersion string
}

func NewCluster() models.Cluster {
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
	os.Remove(c.buildpacks[name])
	delete(c.buildpacks, name)
	return nil
}

func (c *Cluster) NewApp(bpDir, fixtureName string) (models.App, error) {
	tmpPath, err := ioutil.TempDir("/tmp", "cfapi.pack.app.")
	if err != nil {
		return nil, err
	}
	return &App{
		cluster:    c,
		buildpacks: []string{},
		fixture:    filepath.Join(bpDir, "fixtures", fixtureName),
		name:       fixtureName + "_" + cutlass.RandStringRunes(5),
		tmpPath:    tmpPath,
	}, nil
}

func (c *Cluster) HasMultiBuildpack() bool {
	return true
}

func (c *Cluster) buildpack(buildpack string) string {
	if c.buildpacks[buildpack] != "" {
		return c.buildpacks[buildpack]
	}
	return buildpack
}
