package foundation

import (
	"bytes"
	"nginx/int2/cfapi/utils"
	"os"
	"os/exec"
	"path/filepath"
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
	command := exec.Command("cf", "destroy", "-f", a.name)
	command.Stdout = nil
	command.Stderr = &a.Stderr
	return command.Run()
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
