// Command edgeway runs the data plane: it accepts client traffic on entry points
// and forwards it to backend services.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"edgeway/internal/configmodel"
	"edgeway/internal/dataplane/admin"
	"edgeway/internal/dataplane/server"
	"edgeway/internal/logger"
	"edgeway/internal/version"
)

func main() {
	configPath := flag.String("config", "configs/edgeway.yaml", "path to the static config file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, "edgeway:", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	static, err := configmodel.LoadStatic(configPath)
	if err != nil {
		return err
	}

	log, err := logger.New(static.Log.Level, static.Log.Format, os.Stdout)
	if err != nil {
		return err
	}
	log.Info("starting edgeway", "version", version.Version, "commit", version.Commit)

	if f := static.Providers.File; f != nil {
		dyn, err := configmodel.LoadDynamic(f.Path, static.EntryPoints)
		if err != nil {
			return err
		}
		log.Info("dynamic config loaded", "path", f.Path, "routers", len(dyn.Routers), "services", len(dyn.Services))
	}

	// Routing is not implemented yet (roadmap v0.1), so every entry point answers 404.
	var servers []*server.Server
	for name, ep := range static.EntryPoints {
		servers = append(servers, server.New(name, ep.Address, ep.Timeouts, http.NotFoundHandler(), log))
	}
	if static.API.Address != "" {
		servers = append(servers, server.New("api", static.API.Address, configmodel.DefaultTimeouts(), admin.Handler(), log))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return server.Run(ctx, servers, static.LifeCycle.GraceTimeout, log)
}
