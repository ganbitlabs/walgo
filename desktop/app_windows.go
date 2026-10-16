//go:build windows
// +build windows

package main

import (
	"os/exec"
	"syscall"
)

// hideWindow sets Windows-specific flags to hide console window for a command
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// openFileExplorer opens a path in Windows Explorer
func openFileExplorer(path string) error {
	cmd := exec.Command("explorer", path)
	// Don't hide window for explorer - it's supposed to be visible
	return cmd.Start() // Use Start() so we don't wait for explorer to close
}
