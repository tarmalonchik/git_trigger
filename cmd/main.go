package main

import (
	"context"
	"os"

	"github.com/Alan-prog/git_trigger/pkg/commandLine"
	"github.com/Alan-prog/git_trigger/pkg/workers"
	"github.com/sirupsen/logrus"
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
	if err := os.MkdirAll("logs/clone", 0777); err != nil {
		return err
	}
	if err := os.MkdirAll("logs/maker", 0777); err != nil {
		return err
	}
	if err := os.MkdirAll("logs/pull", 0777); err != nil {
		return err
	}

	file, err := os.Create("logs/clone/errors")
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	file, err = os.Create("logs/clone/info")
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	file, err = os.Create("logs/maker/errors")
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	file, err = os.Create("logs/maker/info")
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	file, err = os.Create("logs/pull/errors")
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	file, err = os.Create("logs/pull/info")
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
