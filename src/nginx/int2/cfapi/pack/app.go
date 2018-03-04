package pack

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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

func (a *App) Buildpacks(buildpacks []string) {
	a.buildpacks = buildpacks
}

func (a *App) ConfirmBuildpack(version string) error {
	return nil
}

func (a *App) Stage() error {
	// docker run --rm
	// -v "$(pwd)/1_12_x:/workspace"
	// -v "$(pwd)/out:/out"
	// -v "$(pwd)/buildpacks:/var/lib/buildpacks"
	// packs/cf:build

	fmt.Println("**** Running pack.App.Stage: ", a.tmpPath)

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

	if err := os.RemoveAll(filepath.Join(a.tmpPath, "app")); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(a.tmpPath, "app"), 0755); err != nil {
		return err
	}
	if err := libbuildpack.CopyDirectory(a.fixture, filepath.Join(a.tmpPath, "app")); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(a.tmpPath, "out")); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(a.tmpPath, "out"), 0755); err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}, &container.HostConfig{
		AutoRemove:      true,
		PublishAllPorts: true,
		Binds: []string{
			filepath.Join(a.tmpPath, "app") + ":/workspace",
			filepath.Join(a.tmpPath, "out") + ":/out",
			// -v "$(pwd)/buildpacks:/var/lib/buildpacks"
		},
	}, nil, "")
	if err != nil {
		return err
	}
	containerID := resp.ID

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	fmt.Println("**** ContainerID:", containerID)

	// filter := filters.NewArgs()
	// filter.Add("id", containerID)
	// containers, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filter})
	// if err != nil {
	// 	return err
	// }
	// if len(containers) < 1 {
	// 	return fmt.Errorf("Could not find container with ID: %s", containerID)
	// } else if len(containers) > 1 {
	// 	return fmt.Errorf("Found %d containers with ID: %s", len(containers), containerID)
	// }
	// port := containers[0].Ports[0].PublicPort
	// fmt.Println("**** Port:", port)
	//
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// go func() {
	// 	for _ = range c {
	// 		// sig is a ^C, handle it
	// 		if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
	// 			fmt.Println("Failed to remove container")
	// 		}
	// 	}
	// }()

	out2, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		panic(err)
	}
	io.Copy(&a.Stdout, out2)

	var statusCode int64
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
		panic("Received nil as error on errCh")
	case resp := <-statusCh:
		if resp.Error != nil {
			return fmt.Errorf(resp.Error.Message)
		}
		statusCode = resp.StatusCode
	}

	fmt.Println("**** StatusCode:", statusCode)

	// TODO statusCode should be useful for return success below
	if statusCode != 0 {
		return fmt.Errorf("Docker run %s statusCode %v", imageName, statusCode)
	}
	return nil
}

func (a *App) Run() error {
	return nil
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
func (a *App) Destroy() error {
	return nil
}
func (a *App) GetBody(path string) (string, error) {
	return "", nil
}
func (a *App) Log() string {
	return a.Stdout.String()
}
