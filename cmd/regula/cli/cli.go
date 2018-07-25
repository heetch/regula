package cli

import (
	"context"
	"log"
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
	"github.com/heetch/regula/store"
	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

type Config struct {
	Etcd struct {
		Endpoints string `config:"etcd-endpoints"`
		Namespace string `config:"etcd-namespace,required"`
	}
	Server struct {
		Address      string        `config:"addr"`
		WatchTimeout time.Duration `config:"server-watch-timeout"`
	}
	LogLevel string `config:"log-level"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	cfg.LogLevel = zerolog.DebugLevel.String()
	cfg.Etcd.Endpoints = "127.0.0.1:2379"
	cfg.Server.Address = "0.0.0.0:5331"
	cfg.Server.WatchTimeout = 60 * time.Second

	err := confita.NewLoader(env.NewBackend(), flags.NewBackend()).Load(context.Background(), &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func CreateLogger(level string) zerolog.Logger {
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

func EtcdClient(endpoints string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func CreateServer(cfg *Config, service store.RulesetService, logger zerolog.Logger) *server.Server {
	return server.NewServer(service, server.Config{
		Logger:       &logger,
		WatchTimeout: cfg.Server.WatchTimeout,
	})
}

func RunServer(srv *server.Server, addr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	return srv.Run(ctx, addr)
}
