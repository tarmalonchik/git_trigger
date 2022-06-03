package main

import (
	"context"

	"github.com/Alan-prog/git_trigger/pkg/commandLine"
	"github.com/Alan-prog/git_trigger/pkg/workers"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	repo := "Alan-prog/wireguard_vpn"

	consoleConf, err := commandLine.NewClient(repo)
	if err != nil {
		logrus.Errorf("error init consoleConf: %v", err)
		return
	}

	worker := workers.NewWorker(consoleConf)
	if err := worker.Run(ctx); err != nil {
		logrus.Errorf("error Runner: %v", err)
	}
}
