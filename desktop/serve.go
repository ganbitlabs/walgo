package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ganbitlabs/walgo/pkg/api"
)

const (
	defaultServePort = 1313
	// serveReadyTimeout is how long hugo server gets to start listening.
	serveReadyTimeout = 15 * time.Second
	serveOutputLimit  = 4096
)

// ServeParams holds serve parameters
type ServeParams = api.ServeParams
type ServeResult = api.ServeResult

// hugoServer is the preview server this app started. exited closes once the
// process has been reaped.
type hugoServer struct {
	cmd    *exec.Cmd
	port   int
	exited chan struct{}
}

func (s *hugoServer) url() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

func (s *hugoServer) stop() {
	_ = s.cmd.Process.Kill()
	<-s.exited
}

// Serve builds the site and starts a hugo preview server for it. It returns
// once the server accepts connections, on a free port when the requested one
// is taken.
func (a *App) Serve(params ServeParams) ServeResult {
	sitePath := params.SitePath
	if sitePath == "" {
		return ServeResult{Error: "site path is required"}
	}

	a.StopServe()

	hugoPath, err := api.LookPath("hugo")
	if err != nil {
		return ServeResult{Error: "hugo is not installed or not found in PATH"}
	}

	buildCmd := exec.Command(hugoPath, "--source", sitePath)
	hideWindow(buildCmd)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return ServeResult{Error: fmt.Sprintf("build failed: %s", string(output))}
	}

	port, err := pickPort(params.Port)
	if err != nil {
		return ServeResult{Error: fmt.Sprintf("no free port for the preview server: %v", err)}
	}

	hugoArgs := []string{"server", "--port", strconv.Itoa(port)}
	if params.Drafts {
		hugoArgs = append(hugoArgs, "-D")
	}
	if params.Expired {
		hugoArgs = append(hugoArgs, "-E")
	}
	if params.Future {
		hugoArgs = append(hugoArgs, "-F")
	}

	output := &tailBuffer{}
	cmd := exec.Command(hugoPath, hugoArgs...)
	cmd.Dir = sitePath
	cmd.Stdout = output
	cmd.Stderr = output
	hideWindow(cmd)
	if err := cmd.Start(); err != nil {
		return ServeResult{Error: fmt.Sprintf("failed to start hugo server: %v", err)}
	}

	srv := &hugoServer{cmd: cmd, port: port, exited: make(chan struct{})}
	a.mu.Lock()
	previous := a.serve
	a.serve = srv
	a.mu.Unlock()
	if previous != nil {
		previous.stop()
	}

	go func() {
		_ = cmd.Wait()
		close(srv.exited)
		a.mu.Lock()
		if a.serve == srv {
			a.serve = nil
		}
		a.mu.Unlock()
	}()

	if err := waitListening(srv, serveReadyTimeout); err != nil {
		srv.stop()
		if out := strings.TrimSpace(output.String()); out != "" {
			return ServeResult{Error: fmt.Sprintf("%v:\n%s", err, out)}
		}
		return ServeResult{Error: err.Error()}
	}

	return ServeResult{Success: true, URL: srv.url()}
}

// waitListening returns once srv accepts connections, or an error when it
// exits first or does not listen within timeout.
func waitListening(srv *hugoServer, timeout time.Duration) error {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(srv.port))
	deadline := time.Now().Add(timeout)
	for {
		if conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond); err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("hugo server did not start listening on port %d within %v", srv.port, timeout)
		}
		select {
		case <-srv.exited:
			return fmt.Errorf("hugo server exited before it started serving")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// pickPort returns preferred (1313 when zero) if it is free, otherwise a free
// port chosen by the OS.
func pickPort(preferred int) (int, error) {
	if preferred == 0 {
		preferred = defaultServePort
	}
	if ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(preferred))); err == nil {
		_ = ln.Close()
		return preferred, nil
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

// StopServe stops the preview server this app started.
func (a *App) StopServe() bool {
	a.mu.Lock()
	srv := a.serve
	a.serve = nil
	a.mu.Unlock()

	if srv == nil {
		return false
	}
	srv.stop()
	return true
}

// GetServerURL returns the running preview server's URL, or "" when none runs.
func (a *App) GetServerURL() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.serve == nil {
		return ""
	}
	return a.serve.url()
}

// tailBuffer keeps the last serveOutputLimit bytes written to it, enough to
// explain why hugo server failed to start.
type tailBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	if extra := len(b.buf) - serveOutputLimit; extra > 0 {
		b.buf = append(b.buf[:0], b.buf[extra:]...)
	}
	return len(p), nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}
