package cflocal

import (
	"io/ioutil"
	"nginx/int2/cfapi/models"
	"os"
	"path/filepath"
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

func isDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	return fi.Mode().IsDir(), nil
}

func (c *Cluster) UploadBuildpack(name, version, file string) error {
	if b, err := isDir(file); err == nil && b {
		f, err := ioutil.TempFile("", name)
		if err != nil {
			return err
		}
		f.Close()
		if err := zipit(file, f.Name()); err != nil {
			return err
		}
		file = f.Name()
	}

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
	tmpPath, err := ioutil.TempDir("", "cflocal.app.")
	if err != nil {
		return nil, err
	}
	return &App{
		cluster:    c,
		buildpacks: []string{},
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
