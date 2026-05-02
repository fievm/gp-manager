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
  - Receive `jobShutdownContext` to monitor shutdown signals
  - Jobs should exit gracefully when `jobShutdownContext.Done()` fires
- **Shutdown Jobs**: Register cleanup tasks to run during graceful shutdown
  - Execute after all running jobs have stopped
  - Perfect for closing database connections, flushing logs, etc.
- **Goroutine Management**: Proper synchronization and cleanup of goroutines
- **Error Handling**: Panic recovery and error collection across all jobs
- **Logger Integration**: Built-in Uber Zap logger integration
- **Dual Context System**:
  - `jobShutdownContext`: Signals running jobs to stop gracefully
  - `managerDoneContext`: Signals when entire shutdown is complete

## Shutdown Flow

When a shutdown signal is received (SIGINT, SIGTERM, or parent context cancellation), gp-manager follows this sequence:

1. **Signal Detection**: Receives OS signal or parent context cancellation
2. **Running Jobs Stop** (`jobShutdownContext` cancelled):
   - All running jobs receive the cancellation signal
   - Jobs monitor `jobShutdownContext.Done()` to exit gracefully
   - Each job finishes its current operation cleanly
3. **Cleanup Phase** (Shutdown Jobs Execute):
   - Once all running jobs have stopped, shutdown jobs execute
   - These handle resource cleanup (close DB, flush data, etc.)
4. **Completion** (`managerDoneContext` cancelled):
   - After all cleanup jobs finish, the manager signals completion
   - External code waiting on `manager.Done()` is notified

```
Signal → Running Jobs Stop → Cleanup Jobs Run → Manager Done
```

## Job Types

### Running Jobs
- Type: `func(context.Context) error`
- Receives: `jobShutdownContext`
- Purpose: Long-running processes (HTTP server, message queue consumer, etc.)
- Responsibility: Monitor context and exit gracefully

### Shutdown Jobs
- Type: `func() error`
- Executes: After running jobs have stopped
- Purpose: Cleanup operations (close connections, flush logs, etc.)
- Responsibility: Clean up resources before process exit
| 
| ## Installation

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
	mgr.AddRunningJob(func(jobShutdownCtx context.Context) error {
		for {
			select {
			case <-jobShutdownCtx.Done():
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
- **AddRunningJob(f RunningJob)**: Add a background job that receives jobShutdownContext
- **AddShutdownJob(f ShutdownJob)**: Add a cleanup job for shutdown
- **ShutdownContext() context.Context**: Get the job shutdown context (signals running jobs to stop)
- **Done() <-chan struct{}**: Get channel that closes when manager is completely done

### Job Types

- **RunningJob**: `func(context.Context) error`
  - Receives `jobShutdownContext` as the context parameter
  - Should monitor `jobShutdownContext.Done()` for shutdown signal
  - Examples: HTTP servers, database clients, message queue consumers
  
- **ShutdownJob**: `func() error`
  - Executed after running jobs have stopped
  - Used for cleanup and resource release
  - Examples: Close database connections, flush logs, save state

## Platform Support

- **Unix/Linux/macOS**: Handles SIGINT, SIGTERM, SIGTSTP
- **Windows**: Handles process termination signals

## License

[Add your license here]
