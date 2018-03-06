package foundation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"nginx/int2/cfapi/utils"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tidwall/gjson"
)

type App struct {
	cluster      *Cluster
	buildpacks   []string
	fixture      string
	name         string
	env          map[string]string
	Stdout       bytes.Buffer
	Stderr       bytes.Buffer
	logCmd       *exec.Cmd
	spaceGUID    string
	appGUID      string
	memory       string
	disk         string
	stack        string
	startCommand string
}

func (a *App) Buildpacks(buildpacks []string) {
	a.buildpacks = buildpacks
}
func (a *App) Memory(memory string) {
	a.memory = memory
}
func (a *App) Disk(disk string) {
	a.disk = disk
}
func (a *App) StartCommand(startCommand string) {
	a.startCommand = startCommand
}

func (a *App) ConfirmBuildpack(name string) error {
	return utils.ConfirmBuildpack(a.Log(), a.cluster.defaultBuildpackVersion)
}

func (a *App) PushNoStart() error {
	args := []string{"push", a.name, "--no-start", "-p", a.fixture}
	if a.stack != "" {
		args = append(args, "-s", a.stack)
	}
	if len(a.buildpacks) == 1 {
		args = append(args, "-b", a.buildpacks[len(a.buildpacks)-1])
	}
	if _, err := os.Stat(filepath.Join(a.fixture, "manifest.yml")); err == nil {
		args = append(args, "-f", filepath.Join(a.fixture, "manifest.yml"))
	}
	if a.memory != "" {
		args = append(args, "-m", a.memory)
	}
	if a.disk != "" {
		args = append(args, "-k", a.disk)
	}
	if a.startCommand != "" {
		args = append(args, "-c", a.startCommand)
	}
	command := exec.Command("cf", args...)
	command.Stdout = &a.Stdout
	command.Stderr = &a.Stderr
	if err := command.Run(); err != nil {
		return err
	}

	for k, v := range a.env {
		command := exec.Command("cf", "set-env", a.name, k, v)
		command.Stdout = &a.Stdout
		command.Stderr = &a.Stderr
		if err := command.Run(); err != nil {
			return err
		}
	}

	if a.logCmd == nil {
		a.logCmd = exec.Command("cf", "logs", a.name)
		a.logCmd.Stderr = &a.Stderr
		a.Stdout.Reset()
		a.logCmd.Stdout = &a.Stdout
		if err := a.logCmd.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Push() error {
	if err := a.PushNoStart(); err != nil {
		return err
	}

	var args []string
	if len(a.buildpacks) > 1 {
		args = []string{"v3-push", a.name, "-p", a.fixture}
		for _, buildpack := range a.buildpacks {
			args = append(args, "-b", buildpack)
		}
	} else {
		args = []string{"start", a.name}
	}
	command := exec.Command("cf", args...)
	command.Stdout = nil
	command.Stderr = &a.Stderr
	return command.Run()
}

func (a *App) PushAndConfirm() error {
	if err := a.Push(); err != nil {
		return err
	}
	if err := a.ConfirmBuildpack(""); err != nil {
		return err
	}
	return nil
}

func (a *App) Stop() error {
	command := exec.Command("cf", "stop", a.name)
	command.Stdout = nil
	command.Stderr = &a.Stderr
	return command.Run()
}

func (a *App) Destroy() error {
	command := exec.Command("cf", "delete", "-f", a.name)
	command.Stdout = nil
	command.Stderr = &a.Stderr
	return command.Run()
}

func (a *App) GetUrl(path string) (string, error) {
	guid, err := a.AppGUID()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cf", "curl", "/v2/apps/"+guid+"/summary")
	cmd.Stderr = &a.Stderr
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	schema, found := os.LookupEnv("CUTLASS_SCHEMA")
	if !found {
		schema = "http"
	}
	host := gjson.Get(string(data), "routes.0.host").String()
	domain := gjson.Get(string(data), "routes.0.domain.name").String()
	return fmt.Sprintf("%s://%s.%s%s", schema, host, domain, path), nil
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

func (a *App) SetEnv(key, value string) {
	a.env[key] = value
}

func (a *App) SpaceGUID() (string, error) {
	if a.spaceGUID != "" {
		return a.spaceGUID, nil
	}
	cfHome := os.Getenv("CF_HOME")
	if cfHome == "" {
		cfHome = os.Getenv("HOME")
	}
	bytes, err := ioutil.ReadFile(filepath.Join(cfHome, ".cf", "config.json"))
	if err != nil {
		return "", err
	}
	var config struct {
		SpaceFields struct {
			GUID string
		}
	}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return "", err
	}
	a.spaceGUID = config.SpaceFields.GUID
	return a.spaceGUID, nil
}

func (a *App) AppGUID() (string, error) {
	if a.appGUID != "" {
		return a.appGUID, nil
	}
	guid, err := a.SpaceGUID()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cf", "curl", "/v2/apps?q=space_guid:"+guid+"&q=name:"+a.name)
	cmd.Stderr = &a.Stderr
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var apps struct {
		Resources []struct {
			Metadata struct {
				GUID string `json:"guid"`
			} `json:"metadata"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(bytes, &apps); err != nil {
		return "", err
	}
	if len(apps.Resources) != 1 {
		return "", fmt.Errorf("Expected one app, found %d", len(apps.Resources))
	}
	a.appGUID = apps.Resources[0].Metadata.GUID
	return a.appGUID, nil
}

func (a *App) InstanceStates() ([]string, error) {
	guid, err := a.AppGUID()
	if err != nil {
		return []string{}, err
	}
	cmd := exec.Command("cf", "curl", "/v2/apps/"+guid+"/instances")
	cmd.Stderr = &a.Stderr
	bytes, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	var data map[string]struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return []string{}, err
	}
	var states []string
	for _, value := range data {
		states = append(states, value.State)
	}
	return states, nil
}
