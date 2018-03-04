package cflocal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

type App struct {
	cluster    *Cluster
	Buildpacks []string
	fixture    string
	name       string
	tmpPath    string
	cmd        *exec.Cmd
	port       string
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
}

func (a *App) Stage() error {
	args := []string{"local", "stage", a.name, "-p", a.fixture}
	if len(a.Buildpacks) > 0 {
		for _, b := range a.Buildpacks {
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

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
DONE:
	for {
		select {
		case <-ticker.C:
			if line, err := stdout.ReadString('\n'); err == nil {
				port := regexp.MustCompile(`Running .* on port (\d+)\.\.\.`).FindAllStringSubmatch(line, 1)
				if len(port) == 1 {
					a.port = port[0][1]
					break DONE
				} else {
					return fmt.Errorf("Could not find port in: %s", line)
				}
			}
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Timed out trying to find port")
		}
	}

	for {
		select {
		case <-ticker.C:
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", a.port))
			if err == nil {
				conn.Close()
				return nil
			}
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Timed out waiting to connect to port %s", a.port)
		}
	}
}

func (a *App) ConfirmBuildpack(version string) error {
	// Not needed since always specified // TODO reconsider
	return nil
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
	// TODO
	// Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
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
	client := &http.Client{}
	if headers["NoFollow"] == "true" {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		delete(headers, "NoFollow")
	}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if headers["user"] != "" && headers["password"] != "" {
		req.SetBasicAuth(headers["user"], headers["password"])
		delete(headers, "user")
		delete(headers, "password")
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", map[string][]string{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", map[string][]string{}, err
	}
	resp.Header["StatusCode"] = []string{strconv.Itoa(resp.StatusCode)}
	return string(data), resp.Header, err
}

func (a *App) GetBody(path string) (string, error) {
	body, _, err := a.Get(path, map[string]string{})
	// TODO: Non 200 ??
	// if !(len(headers["StatusCode"]) == 1 && headers["StatusCode"][0] == "200") {
	// 	return "", fmt.Errorf("non 200 status: %v", headers)
	// }
	return body, err
}
