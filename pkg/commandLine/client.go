package commandLine

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	pullString string
	dir        string
}

func NewClient(repo string) (*Client, error) {
	spt := strings.Split(repo, "/")
	if len(spt) != 2 {
		return nil, fmt.Errorf("commandLine.NewClient bad repo")
	}

	c := &Client{}

	c.pullString = "git@github.com:" + repo + ".git"
	c.dir = spt[1]
	return c, nil
}

func (c *Client) Clone(ctx context.Context) error {
	const (
		infoFileName   = "logs/clone/info"
		errorsFileName = "logs/clone/errors"
		lookForString  = "already exists"
	)

	infoFile, err := os.OpenFile(infoFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("commandLine.Clone error making info file: %w", err)
	}
	defer func() { _ = infoFile.Close() }()
	errorsFile, err := os.OpenFile(errorsFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("commandLine.Clone error making errors file: %w", err)
	}
	defer func() { _ = errorsFile.Close() }()

	cmd := exec.CommandContext(ctx, "git", "clone", c.pullString, "--progress")
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Clone error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		data, err := ioutil.ReadFile(errorsFileName)
		if strings.Contains(string(data), lookForString) {
			return nil
		}
		return fmt.Errorf("commandLine.Clone error wait: %w", err)
	}
	return nil
}

func (c *Client) Pull(ctx context.Context) (bool, error) {
	const (
		infoFileName   = "logs/pull/info"
		errorsFileName = "logs/pull/errors"
		lookForString  = "Updating"
	)

	infoFile, err := os.OpenFile(infoFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return false, fmt.Errorf("commandLine.Pull error making info file: %w", err)
	}
	defer func() { _ = infoFile.Close() }()
	errorsFile, err := os.OpenFile(errorsFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return false, fmt.Errorf("commandLine.Pull error making errors file: %w", err)
	}
	defer func() { _ = errorsFile.Close() }()

	cmd := exec.CommandContext(ctx, "git", "pull", "origin", "master", "--progress")
	cmd.Dir = c.dir
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("commandLine.Pull error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return false, fmt.Errorf("commandLine.Pull error wait: %w", err)
	}

	data, err := ioutil.ReadFile(infoFileName)
	if err != nil {
		return false, fmt.Errorf("commandLine.Pull error reading file: %w", err)
	}

	if strings.Contains(string(data), lookForString) {
		return true, nil
	}
	return false, nil
}

func (c *Client) Maker(ctx context.Context, makeCommand string) error {
	const (
		infoFileName   = "logs/maker/info"
		errorsFileName = "logs/maker/errors"
	)

	infoFile, err := os.OpenFile(infoFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("commandLine.Maker error making info file: %w", err)
	}
	defer func() { _ = infoFile.Close() }()
	errorsFile, err := os.OpenFile(errorsFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("commandLine.Maker error making errors file: %w", err)
	}
	defer func() { _ = errorsFile.Close() }()

	cmd := exec.CommandContext(ctx, "make", makeCommand)
	cmd.Dir = c.dir
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Maker error start: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("commandLine.Maker error wait: %w", err)
	}
	return nil
}

func (c *Client) Checkout(ctx context.Context, branchName string) error {
	const (
		lookForString = "did not match any file(s) known to git"
	)
	var (
		info   = bytes.NewBuffer([]byte{})
		errors = bytes.NewBuffer([]byte{})
	)

	cmd := exec.CommandContext(ctx, "git", "checkout", branchName)
	cmd.Dir = c.dir
	cmd.Stdout = info
	cmd.Stderr = errors

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Checkout error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("commandLine.Checkout error wait: %w", err)
	}

	if strings.Contains(errors.String(), lookForString) {
		return nil
	}
	return nil
}
