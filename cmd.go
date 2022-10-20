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
		// read header
		_, err = reader.Read(header)
		// fmt.Println(err, header)

		if err != nil {
			break
		}
		// if err == io.EOF {
		// 	break
		// }
		// if err != nil {
		// 	return decoded, err
		// }

		// typ := binary.BigEndian.Uint16(header[0:1])
		typ := int(header[0])
		size := binary.BigEndian.Uint32(header[4:])
		// fmt.Println(typ, size)

		// read data
		// _, err := io.CopyN(dst, src, n)
		// data := make([]byte, size)
		// _, err = stream.Read(data)
		// if err != nil {
		// break
		// }
		// if err == io.EOF {
		// break
		// }
		// if err != nil {
		// return decoded, err
		// }

		switch typ {
		case 0:
			_, err = io.CopyN(&stdin, reader, int64(size))
			// decoded.stdin += string(data)
		case 1:
			_, err = io.CopyN(&stdout, reader, int64(size))
			// decoded.stdout += string(data)
		case 2:
			_, err = io.CopyN(&stderr, reader, int64(size))
			// decoded.stderr += string(data)
		default:
		}
		if err != nil {
			break
		}

		// decodedstring(data)
		// 0: stdin (is written on stdout)
		// 1: stdout
		// 2: stderr

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
func ExecCmd(container testcontainers.Container, cmd []string, ctx context.Context) (CmdOutput, error) {
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
	// if out, err := ioutil.ReadAll(reader); err == nil {
	// 	output = string(out)
	// }
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
