//go:build windows
// +build windows

package gp_manager

import (
	"os"
	"syscall"
)

var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
