package testcontainers

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

// CmdOutput holds decoded output of a command executed in a container
type CmdOutput struct {
	Stdin  string
	Stdout string
	Stderr string
}

// ReadCmdOutput reads and decodes output of a command executed in a container
func ReadCmdOutput(reader io.Reader) (CmdOutput, error) {
	var err error
	var stdin strings.Builder
	var stdout strings.Builder
	var stderr strings.Builder
	// reference: https://docs.docker.com/engine/api/v1.26/#tag/Container/operation/ContainerAttach
	// header: [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
	header := make([]byte, 8)
	for {
		_, err = reader.Read(header)
		if err != nil {
			break
		}
		typ := int(header[0])
		size := binary.BigEndian.Uint32(header[4:])

		// 0: stdin
		// 1: stdout
		// 2: stderr
		switch typ {
		case 0:
			_, err = io.CopyN(&stdin, reader, int64(size))
		case 1:
			_, err = io.CopyN(&stdout, reader, int64(size))
		case 2:
			_, err = io.CopyN(&stderr, reader, int64(size))
		default:
		}
		if err != nil {
			break
		}

	}
	decoded := CmdOutput{
		Stdin:  stdin.String(),
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err == io.EOF {
		return decoded, nil
	}
	return decoded, err
}

// ExecCmd executes a command in a container and collects its output
func ExecCmd(ctx context.Context, container testcontainers.Container, cmd []string) (CmdOutput, error) {
	var output CmdOutput
	name, _ := container.Name(ctx)
	id := container.GetContainerID()
	exitCode, reader, err := container.Exec(ctx, cmd)
	if err != nil {
		return output, fmt.Errorf(`running:
%s
in %s (%s) failed: %v`, cmd, name, id, err)
	}
	bufReader := bufio.NewReader(reader)
	if decoded, err := ReadCmdOutput(bufReader); err == nil {
		output = decoded
	}
	if exitCode != 0 {
		return output, fmt.Errorf(`running:
%s
in %s (%s) failed:
 => exit code: %d
 => output: %v
 => error: %v`, cmd, name, id, exitCode, output, err)
	}
	return output, nil
}
