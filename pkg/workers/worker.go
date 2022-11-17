package workers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/git_trigger/pkg/commandLine"
)

type Worker struct {
	client      *commandLine.Client
	globalStop  context.CancelFunc
	makeCommand string
	branchName  string

	smallStopFunc context.CancelFunc
	smallStopCtx  context.Context
}

func NewWorker(client *commandLine.Client, makeCommand string, branchName string) *Worker {
	return &Worker{
		client:      client,
		makeCommand: makeCommand,
		branchName:  branchName,
	}
}

func (t *Worker) Run(ctx context.Context) error {
	if err := t.client.Clone(ctx); err != nil {
		return fmt.Errorf("workers.Run error cloning: %w", err)
	}

	if err := t.client.PullAll(ctx); err != nil {
		logrus.Errorf("workers.check error pulling first time: %v", err)
	}

	if err := t.client.Checkout(ctx, t.branchName); err != nil {
		return fmt.Errorf("workers.Run error checkout: %w", err)
	}

	if err := t.client.Maker(ctx, t.makeCommand); err != nil {
		logrus.Errorf("workers.Run error while making first make command: %v", err)
	}

	ctx, t.globalStop = context.WithCancel(ctx)
	t.smallStopCtx, t.smallStopFunc = context.WithCancel(ctx)
	t.smallStopFunc()

	go t.waitForInterruption()

	go t.check(ctx)

	for {
		runtime.GC()
		select {
		case <-ctx.Done():
			logrus.Info("workers.Run successful stop")
			return nil
		case <-t.smallStopCtx.Done():
			time.Sleep(10 * time.Second)
			smallStopCtx, smallStopFunc := context.WithCancel(ctx)
			t.smallStopCtx = smallStopCtx
			t.smallStopFunc = smallStopFunc
			if err := t.client.Maker(t.smallStopCtx, t.makeCommand); err != nil {
				t.smallStopFunc()
				logrus.Errorf("workers.Run error while making make command: %v", err)
			}
		}
	}
}

func (t *Worker) check(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		action, err := t.client.PullBranch(ctx, t.branchName)
		if err != nil {
			logrus.Errorf("workers.check error getting action: %v", err)
			continue
		}
		if action {
			t.smallStopFunc()
		}
	}
}

func (t *Worker) waitForInterruption() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	t.globalStop()
}
