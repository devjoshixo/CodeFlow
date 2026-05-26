package sandbox

import (
	"bytes"
	"codeflow/internal/execution"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func DockerExecutor(ctx context.Context, execution *execution.Execution) (string, string, int, int, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", "", 0, 0, err
	}
	defer cli.Close()

	imageMap := map[string]string{
		"python":     "python:3.11-alpine",
		"go":         "golang:1.21-alpine",
		"javascript": "node:20-alpine",
	}

	cmdMap := map[string][]string{
		"python": {
			"python",
			"-c",
			execution.Code,
		},
		"javascript": {
			"node",
			"-e",
			execution.Code,
		},
		"go": {
			"sh",
			"-c",
			fmt.Sprintf(`cat << 'EOF' > main.go %s EOF go run main.go`, execution.Code),
		},
	}

	imageToPull := imageMap[execution.Language]

	hostConfigs := &container.HostConfig{
		Resources: container.Resources{
			Memory:   64 * 1024 * 1024,
			NanoCPUs: 500000000,
		},
	}

	reader, err := cli.ImagePull(ctx, imageToPull, image.PullOptions{})
	if err != nil {
		return "", "", 0, 0, err
	}
	defer reader.Close()

	io.Copy(io.Discard, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageToPull,
		Cmd:   cmdMap[execution.Language],
	}, hostConfigs, nil, nil, "")

	if err != nil {
		return "", "", 0, 0, err
	}

	startTime := time.Now()

	fmt.Println("execution Id:", resp)

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", 0, 0, err
	}

	exitCode := 0

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	case errCh := <-errCh:
		return "", "", 0, 0, fmt.Errorf("container wait error:%w", errCh)
	}

	read, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "50",
	})
	if err != nil {
		return "", "", 0, 0, err
	}

	defer read.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, read); err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to demux container logs: %w", err)
	}

	cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	//This is the container time that it takes from start to finish
	durationMs := int(time.Since(startTime).Milliseconds())

	return stdout.String(), stderr.String(), exitCode, durationMs, nil
}
