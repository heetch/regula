package main

import (
	"context"
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
	"github.com/heetch/rules-engine/server"
	"github.com/heetch/rules-engine/store/etcd"
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

	srv := server.New(&store, logger)

	runServer(srv, cfg.Server.Address, logger)
}

func loadConfig() *config {
	var cfg config
	cfg.LogLevel = "DEBUG"
	cfg.Etcd.Endpoints = "127.0.0.1:2379"
	cfg.Server.Address = "0.0.0.0:5331"

	err := confita.NewLoader().Load(context.Background(), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}

func createLogger(level string) zerolog.Logger {
	logger := zerolog.New(os.Stdout)

	// zerolog has currently no support for string to level conversion
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for i, lvl := range levels {
		if lvl == level {
			logger = logger.Level(zerolog.Level(i))
		}
	}

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
