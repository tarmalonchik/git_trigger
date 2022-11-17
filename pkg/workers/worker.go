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
	commandLineClient *commandLine.Client

	globalStop    context.CancelFunc
	smallStopFunc context.CancelFunc
	smallStopCtx  context.Context
}

func NewWorker(commandLineClient *commandLine.Client) *Worker {
	return &Worker{
		commandLineClient: commandLineClient,
	}
}

func (t *Worker) Run(ctx context.Context) error {
	if err := t.commandLineClient.Clone(ctx); err != nil {
		return fmt.Errorf("workers.Run error cloning: %w", err)
	}

	if err := t.commandLineClient.PullAll(ctx); err != nil {
		logrus.Errorf("workers.check error pulling first time: %v", err)
	}

	if err := t.commandLineClient.Checkout(ctx); err != nil {
		return fmt.Errorf("workers.Run error checkout: %w", err)
	}

	if err := t.commandLineClient.Maker(ctx); err != nil {
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
			if err := t.commandLineClient.Maker(t.smallStopCtx); err != nil {
				t.smallStopFunc()
				logrus.Errorf("workers.Run error while making make command: %v", err)
			}
		}
	}
}

func (t *Worker) check(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		action, err := t.commandLineClient.PullBranch(ctx)
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
