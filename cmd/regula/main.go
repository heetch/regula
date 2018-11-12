package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula/api/server"
	"github.com/heetch/regula/cmd/regula/cli"
	reghttp "github.com/heetch/regula/http"
	"github.com/heetch/regula/store/etcd"
	"github.com/heetch/regula/ui"
	"github.com/pkg/errors"
)

func main() {
	cfg, err := cli.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := cli.CreateLogger(cfg.LogLevel, os.Stderr)

	err = runServer(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}
}

func runServer(cfg *cli.Config, logger zerolog.Logger) error {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect to etcd cluster")
	}
	defer etcdCli.Close()

	service := etcd.RulesetService{
		Client:    etcdCli,
		Namespace: cfg.Etcd.Namespace,
		Logger:    logger.With().Str("service", "etcd").Logger(),
	}

	srv := cli.NewServer()
	srv.Logger = logger

	var mux http.ServeMux

	mux.Handle("/rulesets/", server.NewHandler(&service, server.Config{
		Timeout:        cfg.Server.Timeout,
		WatchTimeout:   cfg.Server.WatchTimeout,
		WatchCancelCtx: srv.OnShutdownCtx,
	}))

	mux.Handle("/ui/", http.StripPrefix("/ui", ui.NewHandler(logger, cfg.Server.DistPath)))

	// Add middlewares and set the handler to the server
	srv.Handler = reghttp.NewHandler(logger, &mux)

	return srv.Run(context.Background(), cfg.Server.Address)
}
