package mysql

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type MySQLSource struct {
	host     string
	port     int
	user     string
	password string
	database string
	options  string
}

type SourceConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

func NewMySQLSource(config SourceConfig) (*MySQLSource, error) {
	attrs := config.GetOtherAttributes()

	host := attrs["dbhost"]
	if host == "" {
		host = "localhost"
	}

	portStr := attrs["dbport"]
	port := 3306
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	user := attrs["dbuser"]
	if user == "" {
		user = "root"
	}

	password := attrs["dbpass"]
	database := attrs["dbname"]
	if database == "" {
		return nil, fmt.Errorf("MySQL source missing required 'dbname' attribute")
	}

	options := attrs["options"]

	return &MySQLSource{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
		options:  options,
	}, nil
}

func (m *MySQLSource) GetReader() (io.ReadCloser, error) {
	// Create temporary .my.cnf for secure credential passing
	tmpDir := os.TempDir()
	cnfFile := filepath.Join(tmpDir, ".my.cnf."+strconv.Itoa(os.Getpid()))
	defer func() {
		_ = os.Remove(cnfFile)
	}()

	cnfContent := `[client]
host={{ .Host }}
port={{ .Port }}
user={{ .User }}
`
	if m.password != "" {
		cnfContent += `password={{ .Password }}
`
	}

	tmpl, err := template.New("cnf").Parse(cnfContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config template: %v", err)
	}

	cnfData := struct {
		Host     string
		Port     int
		User     string
		Password string
	}{
		Host:     m.host,
		Port:     m.port,
		User:     m.user,
		Password: m.password,
	}

	f, err := os.Create(cnfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := tmpl.Execute(f, cnfData); err != nil {
		return nil, fmt.Errorf("failed to write config: %v", err)
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	if err := os.Chmod(cnfFile, 0600); err != nil {
		return nil, fmt.Errorf("failed to chmod config file: %v", err)
	}

	args := []string{
		"--defaults-file=" + cnfFile,
		m.database,
	}

	if m.options != "" {
		args = append(args, strings.Fields(m.options)...)
	}

	cmd := exec.Command("mysqldump", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mysqldump: %v", err)
	}

	return &cmdReader{cmd: cmd, reader: stdout, cnfFile: cnfFile}, nil
}

type cmdReader struct {
	cmd     *exec.Cmd
	reader  io.ReadCloser
	cnfFile string
}

func (cr *cmdReader) Read(p []byte) (int, error) {
	return cr.reader.Read(p)
}

func (cr *cmdReader) Close() error {
	err1 := cr.reader.Close()
	err2 := cr.cmd.Wait()
	_ = os.Remove(cr.cnfFile)
	if err1 != nil {
		return err1
	}
	return err2
}
