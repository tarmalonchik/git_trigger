package commandLine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	pullString  string
	projectName string
	destPath    string
	makeCommand string
	branchName  string
}

func NewClient(repoName, destPath, makeCommand, branchName string) (*Client, error) {
	spt := strings.Split(repoName, "/")
	if len(spt) != 2 {
		return nil, fmt.Errorf("commandLine.NewClient bad repo")
	}

	return &Client{
		destPath:    destPath,
		pullString:  "git@github.com:" + repoName + ".git",
		projectName: spt[1],
		makeCommand: makeCommand,
		branchName:  branchName,
	}, nil
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
	cmd.Dir = c.destPath
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Clone error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		data, err := os.ReadFile(errorsFileName)
		if strings.Contains(string(data), lookForString) {
			return nil
		}
		return fmt.Errorf("commandLine.Clone error wait: %w", err)
	}
	return nil
}

func (c *Client) PullBranch(ctx context.Context) (bool, error) {
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

	cmd := exec.CommandContext(ctx, "git", "pull", "origin", c.branchName, "--progress")
	cmd.Dir = c.getProjectPath()
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("commandLine.Pull error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return false, fmt.Errorf("commandLine.Pull error wait: %w", err)
	}

	data, err := os.ReadFile(infoFileName)
	if err != nil {
		return false, fmt.Errorf("commandLine.Pull error reading file: %w", err)
	}

	if strings.Contains(string(data), lookForString) {
		return true, nil
	}
	return false, nil
}

func (c *Client) PullAll(ctx context.Context) error {
	const (
		infoFileName   = "logs/pull_all/info"
		errorsFileName = "logs/pull_all/errors"
	)

	infoFile, err := os.OpenFile(infoFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("commandLine.PullAll error making info file: %w", err)
	}
	defer func() { _ = infoFile.Close() }()
	errorsFile, err := os.OpenFile(errorsFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("commandLine.PullAll error making errors file: %w", err)
	}
	defer func() { _ = errorsFile.Close() }()

	cmd := exec.CommandContext(ctx, "git", "pull", "--all", "--progress")
	cmd.Dir = c.getProjectPath()
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.PullAll error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("commandLine.PullAll error wait: %w", err)
	}

	return nil
}

func (c *Client) Maker(ctx context.Context) error {
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

	cmd := exec.CommandContext(ctx, "make", c.makeCommand)
	cmd.Dir = c.getProjectPath()
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Maker error start fullPath = %s: %w", c.projectName, err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("commandLine.Maker error wait: %w", err)
	}
	return nil
}

func (c *Client) Checkout(ctx context.Context) error {
	const (
		lookForString  = "did not match any file(s) known to git"
		infoFileName   = "logs/checkout/info"
		errorsFileName = "logs/checkout/errors"
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

	cmd := exec.CommandContext(ctx, "git", "checkout")
	cmd.Dir = c.getProjectPath()
	cmd.Stdout = infoFile
	cmd.Stderr = errorsFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Checkout error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("commandLine.Checkout error wait: %w", err)
	}

	data, err := os.ReadFile(errorsFileName)
	if err != nil {
		return fmt.Errorf("commandLine.Checkout error reading file: %w", err)
	}

	if strings.Contains(string(data), lookForString) {
		return nil
	}
	return nil
}

func (c *Client) getProjectPath() string {
	if c.destPath[len(c.destPath)-1] == '/' {
		return c.destPath + c.projectName
	}
	return c.destPath + "/" + c.projectName
}
