//go:build !windows
// +build !windows

package main

import (
	"os/exec"
	"runtime"
)

// hideWindow is a no-op on Unix systems (no console window issue)
func hideWindow(cmd *exec.Cmd) {
	// Unix systems don't have the console window issue
}

// openFileExplorer opens a path in the system file manager
func openFileExplorer(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux and others
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start() // Use Start() so we don't wait for file manager to close
}
