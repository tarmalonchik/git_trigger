package workers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Alan-prog/git_trigger/pkg/commandLine"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	client     *commandLine.Client
	globalStop context.CancelFunc

	smallStopFunc context.CancelFunc
	smallStopCtx  context.Context
}

func NewWorker(client *commandLine.Client) *Worker {
	return &Worker{
		client: client,
	}
}

func (t *Worker) Run(ctx context.Context) error {
	ctx, t.globalStop = context.WithCancel(ctx)
	t.smallStopCtx, t.smallStopFunc = context.WithCancel(ctx)
	t.smallStopFunc()

	go t.waitForInterruption()

	go t.check(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("workers.Run successful stop")
			return nil
		case <-t.smallStopCtx.Done():
			time.Sleep(1 * time.Second)
			fmt.Println("trying to make")
			smallStopCtx, smallStopFunc := context.WithCancel(ctx)
			t.smallStopCtx = smallStopCtx
			t.smallStopFunc = smallStopFunc
			fmt.Println("making with small stop")
			if err := t.client.Maker(t.smallStopCtx); err != nil {
				t.smallStopFunc()
				logrus.Errorf("workers.Run error while making make command: %v", err)
			}
			fmt.Println("made")
		}
	}
}

func (t *Worker) check(ctx context.Context) {
	for {
		select {
		case <-time.NewTicker(1 * time.Second).C:
			action, err := t.client.Pull(ctx)
			if err != nil {
				logrus.Errorf("workers.check error getting action: %v", err)
				continue
			}
			if action {
				fmt.Println("was action")
				t.smallStopFunc()
				fmt.Println("was stopped small")
			}
		}
	}
}

func (t *Worker) waitForInterruption() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	t.globalStop()
}
