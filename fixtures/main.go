package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// mkdir -p buildpacks/acf92879c5d3998e550daedb7ab6b513
// (cd buildpacks/acf92879c5d3998e550daedb7ab6b513 && unzip ../../../nginx_buildpack-v0.0.4.20180303225711.zip)
// echo '[{"name":"nginx-buildpack","uri":""}]' > buildpacks/config.json
// docker run --rm -v "$(pwd)/1_12_x:/workspace" -v "$(pwd)/out:/out" -v "$(pwd)/buildpacks:/var/lib/buildpacks" packs/cf:build
// docker run --rm -P -v "$(pwd)/out:/workspace" packs/cf:run

func main() {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageName := "packs/cf:run"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}, &container.HostConfig{
		AutoRemove:      true,
		PublishAllPorts: true,
		Binds: []string{
			"/Users/dgodd/workspace/nginx-buildpack/fixtures/out:/workspace",
		},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)

	filter := filters.NewArgs()
	filter.Add("id", resp.ID)
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filter})
	if err != nil {
		panic(err)
	}
	if len(containers) != 1 {
		panic("Could not find single container with id: " + resp.ID)
	}
	port := containers[0].Ports[0].PublicPort
	fmt.Println("Port:", port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
				fmt.Println("Failed to remove container")
			}
		}
	}()

	out2, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out2)

	// statusCode, err := cli.ContainerWait(ctx, resp.ID)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(statusCode)

	fmt.Println("At end")
}
