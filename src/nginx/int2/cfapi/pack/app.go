package pack

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"nginx/int2/cfapi/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type App struct {
	cluster     *Cluster
	buildpacks  []string
	fixture     string
	name        string
	tmpPath     string
	port        string
	containerID string
	Stdout      bytes.Buffer
	Stderr      bytes.Buffer
}

func (a *App) Buildpacks(buildpacks []string) {
	a.buildpacks = buildpacks
}

func (a *App) ConfirmBuildpack(version string) error {
	// TODO
	return nil
}

func (a *App) setupBuildpackDir(buildpacks []string) error {
	configs := []string{}
	for _, buildpackName := range buildpacks {
		buildpackDir := filepath.Join(a.tmpPath, "buildpacks", fmt.Sprintf("%x", md5.Sum([]byte(buildpackName))))
		if err := os.MkdirAll(buildpackDir, 0755); err != nil {
			return err
		}
		isDir, err := utils.IsDir(a.cluster.buildpack(buildpackName))
		if err != nil {
			return err
		}
		if isDir {
			if err := libbuildpack.CopyDirectory(a.cluster.buildpack(buildpackName), buildpackDir); err != nil {
				return err
			}
		} else {
			if err := libbuildpack.ExtractZip(a.cluster.buildpack(buildpackName), buildpackDir); err != nil {
				return err
			}
		}
		configs = append(configs, fmt.Sprintf(`{"name":"%s","uri":""}`, buildpackName))
	}
	configJson := "[" + strings.Join(configs, ",") + "]"
	return ioutil.WriteFile(filepath.Join(a.tmpPath, "buildpacks", "config.json"), []byte(configJson), 0644)
}

func (a *App) Stage() error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	imageName := "packs/cf:build"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(&a.Stdout, out)

	for _, name := range []string{"app", "out", "buildpacks"} {
		if err := os.RemoveAll(filepath.Join(a.tmpPath, name)); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(a.tmpPath, name), 0755); err != nil {
			return err
		}
	}
	if err := libbuildpack.CopyDirectory(a.fixture, filepath.Join(a.tmpPath, "app")); err != nil {
		return err
	}
	// TODO Get correct dirname and handle set buildpacks
	var additionalFlags []string
	if len(a.buildpacks) > 0 {
		additionalFlags = []string{"-skipDetect=true", "-buildpackOrder=" + strings.Join(a.buildpacks, ",")}
		if err := a.setupBuildpackDir(a.buildpacks); err != nil {
			return err
		}
	} else {
		additionalFlags = []string{}
		if err := a.setupBuildpackDir([]string{a.cluster.defaultBuildpackName}); err != nil {
			return err
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          additionalFlags,
	}, &container.HostConfig{
		AutoRemove:      true,
		PublishAllPorts: true,
		Binds: []string{
			filepath.Join(a.tmpPath, "app") + ":/workspace",
			filepath.Join(a.tmpPath, "out") + ":/out",
			filepath.Join(a.tmpPath, "buildpacks") + ":/var/lib/buildpacks",
		},
	}, nil, "")
	if err != nil {
		return err
	}
	containerID := resp.ID

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	out2, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return err
	}
	io.Copy(&a.Stdout, out2)

	statusCode, err := cli.ContainerWait(ctx, containerID)
	if err != nil {
		return err
	}

	if statusCode != 0 {
		return fmt.Errorf("Docker run %s statusCode %v", imageName, statusCode)
	}
	return nil
}

func (a *App) Run() error {
	if a.containerID != "" {
		return fmt.Errorf("Already running")
	}
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	imageName := "packs/cf:run"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(&a.Stdout, out)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}, &container.HostConfig{
		AutoRemove:      true,
		PublishAllPorts: true,
		Binds: []string{
			filepath.Join(a.tmpPath, "out") + ":/workspace",
		},
	}, nil, a.name)
	if err != nil {
		return err
	}
	a.containerID = resp.ID

	if err := cli.ContainerStart(ctx, a.containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	filter := filters.NewArgs()
	filter.Add("id", a.containerID)
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filter})
	if err != nil {
		return err
	}
	if len(containers) < 1 {
		return fmt.Errorf("Could not find container with ID: %s", a.containerID)
	} else if len(containers) > 1 {
		return fmt.Errorf("Found %d containers with ID: %s", len(containers), a.containerID)
	}
	a.port = fmt.Sprintf("%d", containers[0].Ports[0].PublicPort)

	out2, err := cli.ContainerLogs(ctx, a.containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return err
	}
	go io.Copy(&a.Stdout, out2)

	// TODO duplicate from cflocal
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", a.port))
			if err == nil {
				conn.Close()
				// TODO something better, maybe http connection instead of tcp????
				time.Sleep(1000 * time.Millisecond)
				return nil
			}
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Timed out waiting to connect to port %s", a.port)
		}
	}
}

func (a *App) Push() error {
	if err := a.Stage(); err != nil {
		fmt.Println("***** Stage err:", err)
		fmt.Println(a.Log())
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
	if a.containerID != "" {
		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}
		if err := cli.ContainerRemove(ctx, a.containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
			fmt.Println("Failed to remove container")
		}
		a.containerID = ""
	}
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

func (a *App) Log() string {
	return a.Stdout.String()
}
