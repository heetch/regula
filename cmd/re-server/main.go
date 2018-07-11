package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/flags"
	"github.com/heetch/regula/api/server"
	"github.com/heetch/regula/store/etcd"
	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

type config struct {
	Etcd struct {
		Endpoints string `config:"etcd-endpoints"`
		Namespace string `config:"etcd-namespace,required"`
	}
	Server struct {
		Address string `config:"addr"`
	}
	LogLevel string `config:"log-level"`
}

func main() {
	cfg := loadConfig()
	logger := createLogger(cfg.LogLevel)
	cli := etcdClient(cfg.Etcd.Endpoints)
	defer cli.Close()

	store := etcd.Store{
		Client:    cli,
		Namespace: cfg.Etcd.Namespace,
	}

	srv := createServer(cli, cfg.Etcd.Namespace, &store, logger)

	runServer(srv, cfg.Server.Address, logger)
}

func loadConfig() *config {
	var cfg config
	cfg.LogLevel = zerolog.DebugLevel.String()
	cfg.Etcd.Endpoints = "127.0.0.1:2379"
	cfg.Server.Address = "0.0.0.0:5331"

	err := confita.NewLoader(env.NewBackend(), flags.NewBackend()).Load(context.Background(), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}

func createLogger(level string) zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal(err)
	}

	logger = logger.Level(lvl)

	// pretty print during development
	if isatty.IsTerminal(os.Stdout.Fd()) {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// replace standard logger with zerolog
	log.SetFlags(0)
	log.SetOutput(logger)

	return logger
}

func etcdClient(endpoints string) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

func createServer(cli *clientv3.Client, namespace string, store *etcd.Store, logger zerolog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/health", healthCheckHandler(cli, namespace, logger))
	mux.Handle("/", server.NewHandler(store, server.Config{
		Logger: &logger,
	}))
	return &http.Server{
		Handler: mux,
	}
}

func healthCheckHandler(cli *clientv3.Client, namespace string, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := cli.KV.Get(ctx, namespace, clientv3.WithKeysOnly())
		if err != nil {
			logger.Debug().Str("health", "error").Int("status", http.StatusInternalServerError).Msg("health check request")
			logger.Error().Err(err).Msg("health check error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Debug().Str("health", "ok").Int("status", http.StatusOK).Msg("health check request")
		fmt.Fprintf(w, "OK")
	})
}

func runServer(srv *http.Server, addr string, logger zerolog.Logger) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		logger.Info().Msg("Listening on " + l.Addr().String())
		err := srv.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
