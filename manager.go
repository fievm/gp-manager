package gp_manager

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

// manager represents the graceful server manager interface
var manager *Manager

// startOnce initial graceful manager once
var startOnce = sync.Once{}

type (
	RunningJob  func(context.Context) error
	ShutdownJob func() error
)

// Manager manages the graceful shutdown process
type Manager struct {
	lock                     *sync.RWMutex
	jobShutdownContext       context.Context
	jobShutdownContextCancel context.CancelFunc
	managerDoneContext       context.Context
	managerDoneContextCancel context.CancelFunc
	logger                   *zap.Logger
	runningWaitGroup         *routineGroup
	errors                   []error
	runAtShutdown            []ShutdownJob
}

func (g *Manager) start(ctx context.Context) {
	g.jobShutdownContext, g.jobShutdownContextCancel = context.WithCancel(ctx)
	g.managerDoneContext, g.managerDoneContextCancel = context.WithCancel(context.Background())

	go g.handleSignals(ctx)
}

// doGracefulShutdown graceful shutdown all task
func (g *Manager) doGracefulShutdown() {
	g.jobShutdownContextCancel()
	// doing shutdown job
	for _, f := range g.runAtShutdown {
		func(run ShutdownJob) {
			g.runningWaitGroup.Run(func() {
				g.doShutdownJob(run)
			})
		}(f)
	}
	go func() {
		g.waitForJobs()
		g.lock.Lock()
		g.managerDoneContextCancel()
		g.lock.Unlock()
	}()
}

func (g *Manager) waitForJobs() {
	g.runningWaitGroup.Wait()
}

func (g *Manager) handleSignals(ctx context.Context) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, signals...)
	defer signal.Stop(c)

	pid := syscall.Getpid()
	for {
		select {
		case sig := <-c:
			switch sig {
			case syscall.SIGINT:
				g.logger.Info("Received SIGINT. Shutting down...", zap.Int("pid", pid))
				g.doGracefulShutdown()
				return
			case syscall.SIGTERM:
				g.logger.Info("Received SIGTERM. Shutting down...", zap.Int("pid", pid))
				g.doGracefulShutdown()
				return
			default:
				g.logger.Info("Received signal", zap.Int("pid", pid), zap.String("signal", sig.String()))
			}
		case <-ctx.Done():
			g.logger.Info("Background context for manager closed - Shutting down...", zap.Int("pid", pid), zap.Error(ctx.Err()))
			g.doGracefulShutdown()
			return
		}
	}
}

// doShutdownJob execute shutdown task
func (g *Manager) doShutdownJob(f ShutdownJob) {
	// to handle panic cases from inside the worker
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Errorf("panic in shutdown job: %v", err)
			g.logger.Error(msg.Error())
			g.lock.Lock()
			g.errors = append(g.errors, msg)
			g.lock.Unlock()
		}
	}()
	if err := f(); err != nil {
		g.lock.Lock()
		g.errors = append(g.errors, err)
		g.lock.Unlock()
	}
}

// AddShutdownJob add shutdown task
func (g *Manager) AddShutdownJob(f ShutdownJob) {
	g.lock.Lock()
	g.runAtShutdown = append(g.runAtShutdown, f)
	g.lock.Unlock()
}

// AddRunningJob add running task
func (g *Manager) AddRunningJob(f RunningJob) {
	g.runningWaitGroup.Run(func() {
		// to handle panic cases from inside the worker
		defer func() {
			if err := recover(); err != nil {
				msg := fmt.Errorf("panic in running job: %v", err)
				g.logger.Error(msg.Error())
				g.lock.Lock()
				g.errors = append(g.errors, msg)
				g.lock.Unlock()
			}
		}()
		if err := f(g.jobShutdownContext); err != nil {
			g.lock.Lock()
			g.errors = append(g.errors, err)
			g.lock.Unlock()
		}
	})
}

// Done allows the manager to be viewed as a context.Context.
func (g *Manager) Done() <-chan struct{} {
	return g.managerDoneContext.Done()
}

// ShutdownContext returns a context.Context that is Done at shutdown
func (g *Manager) ShutdownContext() context.Context {
	return g.jobShutdownContext
}

func newManager(ctx context.Context, logger *zap.Logger) *Manager {
	startOnce.Do(func() {

		manager = &Manager{
			lock:             &sync.RWMutex{},
			logger:           logger,
			errors:           make([]error, 0),
			runningWaitGroup: newRoutineGroup(),
		}
		manager.start(ctx)
	})

	return manager
}

// NewManager initial the Manager
func NewManager(ctx context.Context, logger *zap.Logger) *Manager {
	return newManager(ctx, logger)
}

// get the Manager
func GetManager() *Manager {
	if manager == nil {
		panic("please use NewManager to initial the manager first")
	}

	return manager
}
