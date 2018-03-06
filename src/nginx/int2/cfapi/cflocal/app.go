package cflocal

import (
	"bytes"
	"fmt"
	"io"
	"nginx/int2/cfapi/utils"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"
)

type App struct {
	cluster    *Cluster
	buildpacks []string
	fixture    string
	name       string
	tmpPath    string
	cmd        *exec.Cmd
	port       string
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
}

func (a *App) Log() string {
	return a.Stdout.String()
}
func (a *App) Buildpacks(buildpacks []string) {
	a.buildpacks = buildpacks
}

func (a *App) Stage() error {
	args := []string{"local", "stage", a.name, "-p", a.fixture}
	if len(a.buildpacks) > 0 {
		for _, b := range a.buildpacks {
			args = append(args, "-b", a.cluster.buildpack(b))
		}
	} else {
		args = append(args, "-e", "-b", a.cluster.buildpack(a.cluster.defaultBuildpackName))
	}
	cmd := exec.Command("cf", args...)
	cmd.Dir = a.tmpPath
	cmd.Stdout = &a.Stdout
	cmd.Stderr = &a.Stderr
	return cmd.Run()
}

func (a *App) Run() error {
	if a.cmd != nil {
		return fmt.Errorf("Already running")
	}
	a.cmd = exec.Command("cf", "local", "run", a.name)
	a.cmd.Dir = a.tmpPath
	var stdout bytes.Buffer
	a.cmd.Stdout = io.MultiWriter(&a.Stdout, &stdout)
	a.cmd.Stderr = &a.Stderr
	if err := a.cmd.Start(); err != nil {
		return err
	}

	var err error
	a.port, err = findPort(stdout)
	if err != nil {
		return err
	}
	return utils.WaitForHttpPort(a.port)
}

func findPort(stdout bytes.Buffer) (string, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if line, err := stdout.ReadString('\n'); err == nil {
				port := regexp.MustCompile(`Running .* on port (\d+)\.\.\.`).FindAllStringSubmatch(line, 1)
				if len(port) == 1 {
					return port[0][1], nil
				} else {
					return "", fmt.Errorf("Could not find port in: %s", line)
				}
			}
		case <-time.After(3 * time.Second):
			return "", fmt.Errorf("Timed out trying to find port")
		}
	}
}

func (a *App) ConfirmBuildpack(name string) error {
	return utils.ConfirmBuildpack(a.Log(), a.cluster.defaultBuildpackVersion)
}

func (a *App) Push() error {
	if err := a.Stage(); err != nil {
		return err
	}
	return a.Run()
}

func (a *App) PushAndConfirm() error {
	if err := a.Push(); err != nil {
		return err
	}
	if err := a.ConfirmBuildpack(a.cluster.defaultBuildpackName); err != nil {
		return err
	}
	return nil
}

func (a *App) Stop() error {
	if a.cmd == nil {
		return nil
	}
	// if err := a.cmd.Process.Kill(); err != nil {
	// 	return err
	// }
	// if err := a.cmd.Process.Signal(syscall.SIGINT); err != nil {
	// 	return err
	// }
	if err := syscall.Kill(-a.cmd.Process.Pid, syscall.SIGINT); err != nil {
		return err
	}
	a.cmd = nil
	return nil
}

func (a *App) Destroy() error {
	a.Stop()
	return os.RemoveAll(a.tmpPath)
}

func (a *App) GetUrl(path string) (string, error) {
	if a.port == "" {
		return "", fmt.Errorf("app does not have a known running port")
	}
	return fmt.Sprintf("http://localhost:%s%s", a.port, path), nil
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
