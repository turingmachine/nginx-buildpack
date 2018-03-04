package foundation

import (
	"nginx/int2/cfapi/models"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"
)

type Cluster struct {
}

func NewCluster() models.Cluster {
	return &Cluster{}
}

func (c *Cluster) UploadBuildpack(name, version, file string) error {
	return nil
}

func (c *Cluster) DeleteBuildpack(name string) error {
	return nil
}

func (c *Cluster) NewApp(bpDir, fixtureName string) (models.App, error) {
	return &App{
		cluster:    c,
		buildpacks: []string{},
		fixture:    filepath.Join(bpDir, "fixtures", fixtureName),
		name:       fixtureName + "_" + cutlass.RandStringRunes(5),
	}, nil
}

func (c *Cluster) HasMultiBuildpack() bool {
	// TODO
	return true
}
