package encryption

import (
	"fmt"
	"io"
	"os/exec"
)

type GPGEncryptor struct {
	passphrase string
	compressOnly bool
}

func NewGPGEncryptor(passphrase string, compressOnly bool) *GPGEncryptor {
	return &GPGEncryptor{
		passphrase:   passphrase,
		compressOnly: compressOnly,
	}
}

func (g *GPGEncryptor) Encrypt(input io.Reader) (io.ReadCloser, error) {
	args := []string{"--batch", "--quiet"}

	if !g.compressOnly {
		args = append(args, "--symmetric", "--cipher-algo", "AES256")
	}

	if g.passphrase != "" {
		args = append(args, "--passphrase", g.passphrase)
	} else if !g.compressOnly {
		return nil, fmt.Errorf("passphrase required for encryption")
	}

	cmd := exec.Command("gpg", args...)
	cmd.Stdin = input

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start gpg: %v", err)
	}

	return &gpgReader{cmd: cmd, reader: stdout}, nil
}

type gpgReader struct {
	cmd    *exec.Cmd
	reader io.ReadCloser
}

func (gr *gpgReader) Read(p []byte) (int, error) {
	return gr.reader.Read(p)
}

func (gr *gpgReader) Close() error {
	err1 := gr.reader.Close()
	err2 := gr.cmd.Wait()
	if err1 != nil {
		return err1
	}
	return err2
}
