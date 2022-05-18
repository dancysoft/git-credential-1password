package onepassword

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	exec "golang.org/x/sys/execabs"
	"golang.org/x/term"

	"github.com/develerik/git-credential-1password/git"
)

var (
	errNoSessionToken = errors.New("no session token found")
	errNotSupported   = errors.New("os not supported")
)

func getTTYPath() (string, error) {
	ttyPath := ""

	switch runtime.GOOS {
	case "darwin":
	case "linux":
		ttyPath = "/dev/tty"
	case "windows":
		ttyPath = "CON:"
	default:
		return "", errNotSupported
	}

	return ttyPath, nil
}

// Login to 1password.
func (c *Client) Login(timeout uint) error { // nolint:funlen,gocognit,gocyclo // TODO: refactor
	var err error
	c.token, err = git.GetFromCache(c.Account)

	if err != nil {
		return err
	}

	if c.token != "" {
		return nil
	}

	ttyPath, err := getTTYPath()

	if err != nil {
		return err
	}

	tty, err := os.Open(ttyPath)

	if err != nil {
		return err
	}

	defer func() {
		_ = tty.Close()
	}()

	if _, err = fmt.Fprint(os.Stderr, "Enter your 1password master password: "); err != nil {
		return err
	}

	fd := int(tty.Fd())
	pass, err := term.ReadPassword(fd)

	if err != nil {
		return err
	}

	var stdout, stderr, stdin bytes.Buffer

	stdin.Write([]byte(fmt.Sprintf("%s\n", pass)))

	args := []string{"signin", "--cache", "--raw"}

	if isV2() {
		args = append(args, "--account")
	}

	args = append(args, c.Account)

	cmd := exec.Command("op", args...)
	cmd.Stdout = &stdout
	cmd.Stdin = &stdin
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("\n%s", stderr.String()) // nolint:goerr113 // TODO: correctly handle error
	}

	token := stdout.String()
	token = strings.TrimSuffix(token, "\n")

	if token == "" {
		return errNoSessionToken
	}

	c.token = token

	if timeout != 0 {
		return git.StoreInCache(c.Account, c.token, timeout)
	}

	return nil
}
