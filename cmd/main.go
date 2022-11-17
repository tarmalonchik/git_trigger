package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/git_trigger/pkg/commandLine"
	"github.com/tarmalonchik/git_trigger/pkg/workers"
)

const (
	defaultBranchName = "master"
)

func main() {
	var (
		branchName string
	)

	ctx := context.Background()

	args := os.Args
	if len(args) > 4 || len(args) < 3 {
		logrus.Errorf("command format should have 2 or 3 params repoName,makeCommand and branchName (is not requeired default using master)")
		return
	}

	if err := initDirsSystem(); err != nil {
		logrus.Errorf("error while initing dirs: %v", err)
		return
	}

	repo := args[1]
	makeCommand := args[2]
	if len(args) > 3 {
		branchName = args[3]
	}

	if branchName == "" {
		branchName = defaultBranchName
	}

	consoleConf, err := commandLine.NewClient(repo)
	if err != nil {
		logrus.Errorf("error init consoleConf: %v", err)
		return
	}

	worker := workers.NewWorker(consoleConf, makeCommand, branchName)
	if err := worker.Run(ctx); err != nil {
		logrus.Errorf("error Runner: %v", err)
	}
}

func initDirsSystem() error {
	if err := createFolders([]string{
		"logs/clone",
		"logs/maker",
		"logs/pull",
		"logs/checkout",
		"logs/pull_all",
	}); err != nil {
		return fmt.Errorf("error creating fodlers: %w", err)
	}

	if err := createFiles([]string{
		"logs/clone/errors",
		"logs/clone/info",
		"logs/clone/errors",
		"logs/clone/info",
		"logs/maker/errors",
		"logs/maker/info",
		"logs/pull/errors",
		"logs/pull/info",
		"logs/checkout/errors",
		"logs/checkout/info",
		"logs/pull_all/errors",
		"logs/pull_all/info",
	}); err != nil {
		return fmt.Errorf("error creating files: %w", err)
	}
	return nil
}

func createFolders(names []string) error {
	for i := range names {
		if err := os.MkdirAll(names[i], 0777); err != nil {
			return err
		}
	}
	return nil
}

func createFiles(files []string) error {
	for i := range files {
		if err := createFile(files[i]); err != nil {
			return err
		}
	}
	return nil
}

func createFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
