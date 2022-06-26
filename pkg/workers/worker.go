package workers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Alan-prog/git_trigger/pkg/commandLine"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	client      *commandLine.Client
	globalStop  context.CancelFunc
	makeCommand string

	smallStopFunc context.CancelFunc
	smallStopCtx  context.Context
}

func NewWorker(client *commandLine.Client, makeCommand string) *Worker {
	return &Worker{
		client:      client,
		makeCommand: makeCommand,
	}
}

func (t *Worker) Run(ctx context.Context) error {
	if err := t.client.Clone(ctx); err != nil {
		return fmt.Errorf("workers.Run error cloning: %w", err)
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
			time.Sleep(1 * time.Second)
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
		action, err := t.client.Pull(ctx)
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
