package folder

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type FolderSource struct {
	path     string
	excludes []string
}

type SourceConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

func NewFolderSource(config SourceConfig) (*FolderSource, error) {
	attrs := config.GetOtherAttributes()

	path := attrs["path"]
	if path == "" {
		return nil, fmt.Errorf("folder source missing required 'path' attribute")
	}

	excludes := []string{}
	if excludeStr := attrs["excludes"]; excludeStr != "" {
		excludes = strings.Split(excludeStr, ",")
		for i := range excludes {
			excludes[i] = strings.TrimSpace(excludes[i])
		}
	}

	return &FolderSource{
		path:     path,
		excludes: excludes,
	}, nil
}

func (f *FolderSource) GetReader() (io.ReadCloser, error) {
	args := []string{
		"-czf",
		"-",
		"--warning=no-file-changed",
	}

	for _, exclude := range f.excludes {
		args = append(args, "--exclude="+exclude)
	}

	args = append(args, f.path)

	cmd := exec.Command("tar", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start tar: %v", err)
	}

	return &cmdReader{cmd: cmd, reader: stdout}, nil
}

type cmdReader struct {
	cmd    *exec.Cmd
	reader io.ReadCloser
}

func (cr *cmdReader) Read(p []byte) (int, error) {
	return cr.reader.Read(p)
}

func (cr *cmdReader) Close() error {
	err1 := cr.reader.Close()
	err2 := cr.cmd.Wait()
	if err1 != nil {
		return err1
	}
	return err2
}
