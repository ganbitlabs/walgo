package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestFakeHugo is not a test. fakeHugo runs the test binary as hugo, and this
// function plays the part: builds succeed, and "server" either exits at once
// (FAKE_HUGO_MODE=die) or listens on its --port until killed.
func TestFakeHugo(t *testing.T) {
	if os.Getenv("FAKE_HUGO") != "1" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 || args[0] != "server" {
		os.Exit(0)
	}
	if os.Getenv("FAKE_HUGO_MODE") == "die" {
		fmt.Fprintln(os.Stderr, "Error: module \"missing-theme\" not found")
		os.Exit(1)
	}
	port := ""
	for i := range args {
		if args[i] == "--port" && i+1 < len(args) {
			port = args[i+1]
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			os.Exit(1)
		}
		_ = conn.Close()
	}
}

func fakeHugo(t *testing.T, mode string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake hugo is a shell script")
	}
	bin := t.TempDir()
	script := fmt.Sprintf("#!/bin/sh\nexec %q -test.run=^TestFakeHugo$ -- \"$@\"\n", os.Args[0])
	if err := os.WriteFile(filepath.Join(bin, "hugo"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("FAKE_HUGO", "1")
	t.Setenv("FAKE_HUGO_MODE", mode)
}

func TestServeWaitsUntilListening(t *testing.T) {
	fakeHugo(t, "listen")
	app := NewApp()
	t.Cleanup(func() { app.StopServe() })

	result := app.Serve(ServeParams{SitePath: t.TempDir()})
	if !result.Success {
		t.Fatalf("Serve failed: %s", result.Error)
	}
	port := strings.TrimPrefix(result.URL, "http://localhost:")
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		t.Fatalf("Serve reported success but %s refuses connections: %v", result.URL, err)
	}
	_ = conn.Close()
	if got := app.GetServerURL(); got != result.URL {
		t.Errorf("GetServerURL = %q, want %q", got, result.URL)
	}
	if !app.StopServe() {
		t.Error("StopServe found no server")
	}
	if got := app.GetServerURL(); got != "" {
		t.Errorf("GetServerURL after stop = %q, want empty", got)
	}
}

func TestServeReportsHugoExit(t *testing.T) {
	fakeHugo(t, "die")
	app := NewApp()

	result := app.Serve(ServeParams{SitePath: t.TempDir()})
	if result.Success {
		t.Fatal("Serve reported success although hugo server exited")
	}
	if !strings.Contains(result.Error, "missing-theme") {
		t.Errorf("error = %q, want hugo's own output", result.Error)
	}
	if got := app.GetServerURL(); got != "" {
		t.Errorf("GetServerURL = %q, want empty", got)
	}
}

func TestServeMovesOffATakenPort(t *testing.T) {
	fakeHugo(t, "listen")
	taken, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer taken.Close()
	takenPort := taken.Addr().(*net.TCPAddr).Port

	app := NewApp()
	t.Cleanup(func() { app.StopServe() })
	result := app.Serve(ServeParams{SitePath: t.TempDir(), Port: takenPort})
	if !result.Success {
		t.Fatalf("Serve failed: %s", result.Error)
	}
	if result.URL == "http://localhost:"+strconv.Itoa(takenPort) {
		t.Errorf("Serve used the taken port %d", takenPort)
	}
}

func TestServeClearsStateWhenHugoDiesLater(t *testing.T) {
	fakeHugo(t, "listen")
	app := NewApp()
	result := app.Serve(ServeParams{SitePath: t.TempDir()})
	if !result.Success {
		t.Fatalf("Serve failed: %s", result.Error)
	}

	app.mu.Lock()
	proc := app.serve.cmd.Process
	app.mu.Unlock()
	_ = proc.Kill()

	deadline := time.Now().Add(5 * time.Second)
	for app.GetServerURL() != "" {
		if time.Now().After(deadline) {
			t.Fatal("server URL still reported after hugo exited")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if app.StopServe() {
		t.Error("StopServe reported a server that had already exited")
	}
}
