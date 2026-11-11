package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"edgeway/internal/configmodel"
)

// Server is a named HTTP listener: an entry point or the admin API.
type Server struct {
	Name string
	http *http.Server
}

func New(name, addr string, t configmodel.Timeouts, h http.Handler, log *slog.Logger) *Server {
	return &Server{
		Name: name,
		http: &http.Server{
			Addr:              addr,
			Handler:           h,
			ReadHeaderTimeout: t.ReadHeader,
			ReadTimeout:       t.Read,
			WriteTimeout:      t.Write,
			IdleTimeout:       t.Idle,
			MaxHeaderBytes:    1 << 20,
			ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
		},
	}
}

// Run serves all servers until ctx is cancelled or one of them fails, then
// shuts every server down, giving in-flight requests up to grace to finish.
func Run(ctx context.Context, servers []*Server, grace time.Duration, log *slog.Logger) error {
	// Bind every address before serving so a port conflict fails startup cleanly.
	listeners := make([]net.Listener, 0, len(servers))
	for _, s := range servers {
		ln, err := net.Listen("tcp", s.http.Addr)
		if err != nil {
			for _, l := range listeners {
				l.Close()
			}
			return fmt.Errorf("listen %s on %s: %w", s.Name, s.http.Addr, err)
		}
		listeners = append(listeners, ln)
	}

	errCh := make(chan error, len(servers))
	for i, s := range servers {
		log.Info("listening", "server", s.Name, "address", listeners[i].Addr().String())
		go func() {
			if err := s.http.Serve(listeners[i]); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("serve %s: %w", s.Name, err)
			}
		}()
	}

	var runErr error
	select {
	case <-ctx.Done():
		log.Info("shutting down", "grace", grace.String())
	case runErr = <-errCh:
		log.Error("server failed, shutting down", "error", runErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()

	errs := make([]error, len(servers))
	var wg sync.WaitGroup
	for i, s := range servers {
		wg.Go(func() {
			if err := s.http.Shutdown(shutdownCtx); err != nil {
				errs[i] = fmt.Errorf("shutdown %s: %w", s.Name, err)
			}
		})
	}
	wg.Wait()

	return errors.Join(append(errs, runErr)...)
}
