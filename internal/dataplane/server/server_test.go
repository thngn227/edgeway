package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"edgeway/internal/configmodel"
)

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().String()
}

func waitListening(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if conn, err := net.Dial("tcp", addr); err == nil {
			conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server on %s never started listening", addr)
}

func TestRunDrainsInFlightRequestsOnShutdown(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	addr := freeAddr(t)

	started := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		time.Sleep(200 * time.Millisecond)
		io.WriteString(w, "done")
	})

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- Run(ctx, []*Server{New("web", addr, configmodel.DefaultTimeouts(), h, log)}, 2*time.Second, log)
	}()
	waitListening(t, addr)

	type result struct {
		body string
		err  error
	}
	resCh := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + addr)
		if err != nil {
			resCh <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		resCh <- result{body: string(b), err: err}
	}()

	<-started
	cancel()

	res := <-resCh
	if res.err != nil || res.body != "done" {
		t.Fatalf("in-flight request: body=%q err=%v, want body \"done\"", res.body, res.err)
	}
	if err := <-runErr; err != nil {
		t.Fatalf("Run returned %v, want nil", err)
	}
}

func TestRunFailsWhenPortIsTaken(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()

	srv := New("web", busy.Addr().String(), configmodel.DefaultTimeouts(), http.NotFoundHandler(), log)
	err = Run(context.Background(), []*Server{srv}, time.Second, log)
	if err == nil || !strings.Contains(err.Error(), "listen web") {
		t.Fatalf("want listen error, got %v", err)
	}
}
