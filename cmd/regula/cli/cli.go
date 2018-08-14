package cli

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/flags"
	"github.com/heetch/regula/api/server"
	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

// Config holds the server configuration.
type Config struct {
	Etcd struct {
		Endpoints []string `config:"etcd-endpoints"`
		Namespace string   `config:"etcd-namespace,required"`
	}
	Server struct {
		Address      string        `config:"addr"`
		Timeout      time.Duration `config:"server-timeout"`
		WatchTimeout time.Duration `config:"server-watch-timeout"`
	}
	LogLevel string `config:"log-level"`
}

// LoadConfig loads the configuration from the environment or command line flags.
func LoadConfig() (*Config, error) {
	var cfg Config
	cfg.LogLevel = zerolog.DebugLevel.String()
	cfg.Etcd.Endpoints = []string{"127.0.0.1:2379"}
	cfg.Server.Address = "0.0.0.0:5331"
	cfg.Server.Timeout = 5 * time.Second
	cfg.Server.WatchTimeout = 30 * time.Second

	err := confita.NewLoader(env.NewBackend(), flags.NewBackend()).Load(context.Background(), &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// CreateLogger returns a configured logger.
func CreateLogger(level string, w io.Writer) zerolog.Logger {
	logger := zerolog.New(w).With().Timestamp().Logger()

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal(err)
	}

	logger = logger.Level(lvl)

	// pretty print during development
	if f, ok := w.(*os.File); ok {
		if isatty.IsTerminal(f.Fd()) {
			logger = logger.Output(zerolog.ConsoleWriter{Out: f})
		}
	}

	// replace standard logger with zerolog
	log.SetFlags(0)
	log.SetOutput(logger)

	return logger
}

// RunServer runs the server and listens to SIGINT and SIGTERM
// to stop the server gracefully.
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
