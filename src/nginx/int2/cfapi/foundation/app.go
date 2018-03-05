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
	cluster    *Cluster
	buildpacks []string
	fixture    string
	name       string
	env        map[string]string
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	logCmd     *exec.Cmd
}

func (a *App) Buildpacks(buildpacks []string) {
	a.buildpacks = buildpacks
}

func (a *App) ConfirmBuildpack(version string) error {
	// TODO
	return nil
}

func (a *App) PushNoStart() error {
	args := []string{"push", a.name, "--no-start", "-p", a.fixture}
	// if a.Stack != "" {
	// 	args = append(args, "-s", a.Stack)
	// }
	if len(a.buildpacks) == 1 {
		args = append(args, "-b", a.buildpacks[len(a.buildpacks)-1])
	}
	if _, err := os.Stat(filepath.Join(a.fixture, "manifest.yml")); err == nil {
		args = append(args, "-f", filepath.Join(a.fixture, "manifest.yml"))
	}
	// if a.Memory != "" {
	// 	args = append(args, "-m", a.Memory)
	// }
	// if a.Disk != "" {
	// 	args = append(args, "-k", a.Disk)
	// }
	// if a.StartCommand != "" {
	// 	args = append(args, "-c", a.StartCommand)
	// }
	// if a.StartCommand != "" {
	// 	args = append(args, "-c", a.StartCommand)
	// }
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
		// TODO clear a.Stdout
		// a.Stdout = bytes.NewBuffer(nil)
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
	// TODO
	// Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
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
	cfHome := os.Getenv("CF_HOME")
	if cfHome == "" {
		cfHome = os.Getenv("HOME")
	}
	bytes, err := ioutil.ReadFile(filepath.Join(cfHome, ".cf", "config.json"))
	if err != nil {
		return "", err
	}
	var config cfConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return "", err
	}
	return config.SpaceFields.GUID, nil
}

func (a *App) AppGUID() (string, error) {
	if a.appGUID != "" {
		return a.appGUID, nil
	}
	guid, err := a.SpaceGUID()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cf", "curl", "/v2/apps?q=space_guid:"+guid+"&q=name:"+a.Name)
	cmd.Stderr = DefaultStdoutStderr
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var apps cfApps
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
	cmd.Stderr = DefaultStdoutStderr
	bytes, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	var data map[string]cfInstance
	if err := json.Unmarshal(bytes, &data); err != nil {
		return []string{}, err
	}
	var states []string
	for _, value := range data {
		states = append(states, value.State)
	}
	return states, nil
}
