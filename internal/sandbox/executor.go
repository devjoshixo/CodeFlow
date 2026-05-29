package sandbox

import (
	"codeflow/internal/execution"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/redis/go-redis/v9"
)

type OutputChunk struct {
	Sequence int
	Stream   string
	Data     string
}

func DockerExecutor(ctx context.Context, redis *redis.Client, execution *execution.Execution) (string, string, int, int, error) {

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
	read, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return "", "", 0, 0, err
	}

	defer read.Close()

	sequence := 0
	channel := fmt.Sprintf("execution:%s:output", execution.ID)

	publish := func(stream string, data []byte) error {
		chunk := OutputChunk{Sequence: sequence, Stream: stream, Data: string(data)}
		payload, err := json.Marshal(chunk)
		if err != nil {
			return err
		}
		if err := redis.Publish(ctx, channel, payload).Err(); err != nil {
			return err
		}
		sequence++
		return nil
	}

	for {
		streamType, data, err := readDockerFrame(read)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", "", 0, 0, fmt.Errorf("error in reading docker logs:%w", err)
		}

		if len(data) == 0 {
			continue
		}

		if err := publish(streamType, data); err != nil {
			return "", "", 0, 0, err
		}
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	case errCh := <-errCh:
		return "", "", 0, 0, fmt.Errorf("container wait error:%w", errCh)
	}

	cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	//This is the container time that it takes from start to finish
	durationMs := int(time.Since(startTime).Milliseconds())

	return "", "", exitCode, durationMs, nil
}

// reads docker log in chunks for live redis update
func readDockerFrame(read io.Reader) (string, []byte, error) {
	header := make([]byte, 8)

	_, err := io.ReadFull(read, header)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't read the stream:%w", err)
	}

	streamType := ""

	switch header[0] {
	case 1:
		streamType = "stdout"
	case 2:
		streamType = "stderr"
	default:
		return "", nil, fmt.Errorf("couldn't determine the stream type:%w", err)
	}

	frameSize := binary.BigEndian.Uint32(header[4:8])

	data := make([]byte, frameSize)

	_, err = io.ReadFull(read, data)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't read the data bytes:%w", err)
	}

	return streamType, data, nil
}

// func readDockerFrame(read io.Reader) (string, []byte, error) {
// 	header := make([]byte, 8)
// 	// Read 8 bytes
// 	// If error, return it

// 	streamType := "" // Check header[0]: 1=stdout, 2=stderr

// 	// Get frame size from header[4:8]
// 	// frameSize := binary.BigEndian.Uint32(header[4:8])

// 	data := make([]byte, frameSize)
// 	// Read frameSize bytes into data
// 	// If error, return it

// 	return streamType, data, nil
// }
