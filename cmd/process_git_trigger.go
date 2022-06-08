package main

import (
	"context"
	"os"

	"github.com/Alan-prog/git_trigger/pkg/commandLine"
	"github.com/Alan-prog/git_trigger/pkg/workers"
	"github.com/sirupsen/logrus"
)

func main() {
	args := os.Args
	if len(args) != 3 {
		logrus.Errorf("command format should have 2 params repoName and makeCommand")
		return
	}

	ctx := context.Background()

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
