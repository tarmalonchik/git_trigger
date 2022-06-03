package commandLine

import (
	"bytes"
	"context"
	"fmt"
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
	var (
		out    bytes.Buffer
		outErr bytes.Buffer
	)
	cmd := exec.CommandContext(ctx, "git", "clone", c.pullString)
	cmd.Stdout = &out
	cmd.Stderr = &outErr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Clone error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if strings.Contains(outErr.String(), "already exists") {
			return nil
		}
		return fmt.Errorf("commandLine.Clone error wait: %w", err)
	}
	return nil
}

func (c *Client) Pull(ctx context.Context) (bool, error) {
	var (
		out bytes.Buffer
	)

	cmd := exec.CommandContext(ctx, "git", "pull", "origin", "master", "--progress")
	cmd.Dir = c.dir
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("commandLine.Pull error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return false, fmt.Errorf("commandLine.Pull error wait: %w", err)
	}

	if strings.Contains(out.String(), "Updating") {
		return true, nil
	}
	return false, nil
}

func (c *Client) Maker(ctx context.Context) error {
	var (
		out bytes.Buffer
	)

	cmd := exec.CommandContext(ctx, "make", "master")
	cmd.Dir = c.dir
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("commandLine.Maker error start: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("commandLine.Maker error wait: %w", err)
	}
	fmt.Println(out.String())

	return nil
}
