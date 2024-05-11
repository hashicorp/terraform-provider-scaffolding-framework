package cli

import (
	"bytes"
	"errors"
	"os/exec"
)

type Cli struct {
	token      string
	server_url string
	binary     string
}

func NewCli(token string, server_url string) *Cli {
	return &Cli{
		token:      token,
		server_url: server_url,
		binary:     "bws",
	}
}

func (c *Cli) ExecuteCommand(args ...string) ([]byte, error) {
	arguments := []string{"--access-token", c.token}

	if c.server_url != "" {
		arguments = append(arguments, "--server-url", c.server_url)
	}

	arguments = append(arguments, args...)

	// var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(c.binary, arguments...)
	// cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	res, err := cmd.Output()

	if err != nil {
		return nil, errors.New(stderr.String())
	}

	return res, nil
}
