package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
)

var ErrSignalled = errors.New("process signalled for shutdown")

// ProcessSpawnFunc is used to start a long running process.
// On exit it cancels all other long running processes sharing the same context.
type ProcessSpawnFunc func(cb func(ctx context.Context) error, procName string)

type Controller struct {
	wg *sync.WaitGroup
}

func NewController() *Controller {
	return &Controller{
		wg: &sync.WaitGroup{},
	}
}

// Start creates an app context that can be used to spawn a group of LRPs.
//
// If one LRP exits then all other LRPs are stopped.
func (c *Controller) Start() (context.Context, ProcessSpawnFunc) {
	ctx, stopFn := context.WithCancelCause(context.Background())
	log := logger.GetLogger()

	procSpawnFn := func(cb func(ctx context.Context) error, procName string) {
		c.wg.Add(1)

		go func() {
			log.Info(fmt.Sprintf("process started - %s", procName))
			defer func() {
				log.Info("stopping process")
				c.wg.Done()
			}()

			err := cb(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Info(fmt.Sprintf("process stopped on context cancelled - %s", procName))
				} else {
					log.Error(fmt.Errorf("failed process - %s", procName))
					log.Error(err)
				}
			}

			stopFn(err)
		}()
	}

	spawnSignalListener(procSpawnFn)

	return ctx, procSpawnFn
}

// LRP that exits on an os signal or context cancenlled, stopping all other LRPs.
func spawnSignalListener(spawnFn ProcessSpawnFunc) {
	spawnFn(func(ctx context.Context) error {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case s := <-signalChan:
			return fmt.Errorf("received %s: %w", s.String(), ErrSignalled)
		case <-ctx.Done():
		}

		return nil
	}, "SignalListener")
}

// Blocks until all LRPs are stopped.
func (c *Controller) Wait() {
	c.wg.Wait()
}
