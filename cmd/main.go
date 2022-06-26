package main

import (
	"context"
	"os"

	"github.com/Alan-prog/git_trigger/pkg/commandLine"
	"github.com/Alan-prog/git_trigger/pkg/workers"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	logrus.Printf("here is some shit")

	args := os.Args
	if len(args) != 3 {
		logrus.Errorf("command format should have 2 params repoName and makeCommand")
		return
	}

	if err := initDirsSystem(); err != nil {
		logrus.Errorf("error while initing dirs: %v", err)
		return
	}

	repo := args[1]
	makeCommand := args[2]

	consoleConf, err := commandLine.NewClient(repo)
	if err != nil {
		logrus.Errorf("error init consoleConf: %v", err)
		return
	}

	worker := workers.NewWorker(consoleConf, makeCommand)
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
