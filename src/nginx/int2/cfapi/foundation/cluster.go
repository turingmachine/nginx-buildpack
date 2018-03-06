package foundation

import (
	"fmt"
	"nginx/int2/cfapi/models"
	"os/exec"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack/cutlass"
)

type Cluster struct {
	defaultBuildpackVersion string
}

func NewCluster() models.Cluster {
	return &Cluster{}
}

func (c *Cluster) UploadBuildpack(name, version, file string) error {
	command := exec.Command("cf", "update-buildpack", name, "-p", file)
	out1, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	command = exec.Command("cf", "create-buildpack", name, file, "100", "--enable")
	out2, err := command.CombinedOutput()
	if err != nil {
		fmt.Println(string(out1), "\n", string(out2))
	}
	if c.defaultBuildpackVersion == "" {
		c.defaultBuildpackVersion = version
	}
	return err
}

func (c *Cluster) DeleteBuildpack(name string) error {
	command := exec.Command("cf", "delete-buildpack", "-f", name)
	_, err := command.CombinedOutput()
	return err
}

func (c *Cluster) NewApp(bpDir, fixtureName string) (models.App, error) {
	return &App{
		cluster:    c,
		buildpacks: []string{},
		env:        map[string]string{},
		fixture:    filepath.Join(bpDir, "fixtures", fixtureName),
		name:       fixtureName + "_" + cutlass.RandStringRunes(5),
	}, nil
}

func ApiGreaterThan(version string) (bool, error) {
	apiVersionString, err := cutlass.ApiVersion()
	if err != nil {
		return false, err
	}
	apiVersion, err := semver.Make(apiVersionString)
	if err != nil {
		return false, err
	}
	reqVersion, err := semver.ParseRange(">= " + version)
	if err != nil {
		return false, err
	}
	return reqVersion(apiVersion), nil
}

func (c *Cluster) HasTask() bool {
	b, err := ApiGreaterThan("2.75.0")
	if err != nil {
		panic("Could not determine if foundation has task support")
	}
	return b
}
func (c *Cluster) HasMultiBuildpack() bool {
	b, err := ApiGreaterThan("2.90.0")
	if err != nil {
		panic("Could not determine if foundation has multi buildpack support")
	}
	return b
}
