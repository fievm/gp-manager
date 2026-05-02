# gp-manager

A graceful Go process manager for handling graceful shutdowns, background jobs, and signal management.

## Overview

`gp-manager` is a lightweight Go library that provides a robust way to manage application lifecycle, including:

- **Graceful Shutdown**: Cleanly shutdown all running jobs and services
- **Background Job Management**: Run and manage background goroutines
- **Signal Handling**: Handle OS signals (SIGINT, SIGTERM, SIGTSTP) gracefully
- **Shutdown Hooks**: Execute cleanup jobs before complete shutdown
- **Cross-Platform Support**: Works on Unix/Linux/macOS and Windows
- **Context-Based Cancellation**: Control job execution via context

## Features

- **Manager**: Central component for managing application lifecycle
- **Running Jobs**: Execute background tasks that respect shutdown signals
- **Shutdown Jobs**: Register cleanup tasks to run during graceful shutdown
- **Goroutine Management**: Proper synchronization and cleanup of goroutines
- **Error Handling**: Panic recovery and error collection across all jobs
- **Logger Integration**: Built-in Uber Zap logger integration

## Installation

```bash
go get github.com/yourusername/gp-manager
```

## Usage

### Basic Example

```go
package main

import (
	"context"
	"log"
	
	"github.com/yourusername/gp-manager"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Initialize the manager
	mgr := gp_manager.NewManager(ctx, logger)
	
	// Add a running job
	mgr.AddRunningJob(func(shutdownCtx context.Context) error {
		for {
			select {
			case <-shutdownCtx.Done():
				logger.Info("Job shutting down")
				return nil
			default:
				// Do work here
			}
		}
	})
	
	// Add a shutdown job
	mgr.AddShutdownJob(func() error {
		logger.Info("Cleaning up resources")
		return nil
	})
	
	// Wait for shutdown signal
	<-mgr.Done()
	logger.Info("Application stopped")
}
```

## API Reference

### Manager

- **NewManager(ctx context.Context, logger *zap.Logger) *Manager**: Initialize a new manager
- **GetManager() *Manager**: Get the singleton manager instance
- **AddRunningJob(f RunningJob)**: Add a background job
- **AddShutdownJob(f ShutdownJob)**: Add a cleanup job for shutdown
- **ShutdownContext() context.Context**: Get the shutdown context
- **Done() <-chan struct{}**: Get channel that closes when manager is done

### Job Types

- **RunningJob**: `func(context.Context) error` - Long-running background job
- **ShutdownJob**: `func() error` - Cleanup job executed at shutdown

## Platform Support

- **Unix/Linux/macOS**: Handles SIGINT, SIGTERM, SIGTSTP
- **Windows**: Handles process termination signals

## License

[Add your license here]
