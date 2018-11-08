// Package cli provides types and functions that are specific to the Regula program.
// These helpers are provided separately from the main package to be able to customize
// Regula server's behaviour to create custom binaries.
package cli

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/flags"
	isatty "github.com/mattn/go-isatty"
	"github.com/pkg/errors"
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
		DistPath     string        `config:"server-dist-path"`
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
	cfg.Server.DistPath = "dist"

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

// Server is a ready to use HTTP server that shuts downs automatically
// when receiving SIGINT or SIGTERM signals.
type Server struct {
	// The http handler this server must serve.
	Handler http.Handler
	// OnShutdownCtx is canceled when the server is shutting down.
	OnShutdownCtx context.Context
	Logger        zerolog.Logger
	srv           http.Server
	cancelFn      func()
}

// NewServer creates an HTTP Server
func NewServer() *Server {
	var s Server

	s.OnShutdownCtx, s.cancelFn = context.WithCancel(context.Background())

	s.srv.RegisterOnShutdown(s.cancelFn)

	return &s
}

// Run runs the http server on the chosen address. The server will shutdown automatically
// if the SIGINT or SIGTERM signal are catched. It can also be closed manually by canceling the
// given context.
func (s *Server) Run(ctx context.Context, addr string) error {
	s.srv.Handler = s.Handler

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		s.Logger.Info().Msg("Listening on " + l.Addr().String())
		err = s.srv.Serve(l)
		if err != http.ErrServerClosed {
			panic(err)
		}
	}()

	select {
	case sig := <-quit:
		s.Logger.Debug().Msgf("received %s signal, shutting down...", sig)
	case <-ctx.Done():
		s.Logger.Debug().Msgf("%s, shutting down...", ctx.Err())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.srv.Shutdown(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to shutdown server gracefully")
	}

	s.Logger.Debug().Msg("server shutdown successfully")

	return nil
}
