package postgresql

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

type PostgreSQLSource struct {
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

func NewPostgreSQLSource(config SourceConfig) (*PostgreSQLSource, error) {
	attrs := config.GetOtherAttributes()

	host := attrs["dbhost"]
	if host == "" {
		host = "localhost"
	}

	portStr := attrs["dbport"]
	port := 5432
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	user := attrs["dbuser"]
	if user == "" {
		user = "postgres"
	}

	password := attrs["dbpass"]
	database := attrs["dbname"]
	if database == "" {
		return nil, fmt.Errorf("PostgreSQL source missing required 'dbname' attribute")
	}

	options := attrs["options"]

	return &PostgreSQLSource{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
		options:  options,
	}, nil
}

func (p *PostgreSQLSource) GetReader() (io.ReadCloser, error) {
	args := []string{
		"-h", p.host,
		"-p", strconv.Itoa(p.port),
		"-U", p.user,
		"-d", p.database,
		"--no-password",
	}

	if p.options != "" {
		args = append(args, p.options)
	}

	cmd := exec.Command("pg_dump", args...)

	if p.password != "" {
		cmd.Env = append(cmd.Environ(), "PGPASSWORD="+p.password)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start pg_dump: %v", err)
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
