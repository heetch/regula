// Package cli provides types and functions that are specific to the Regula program.
// These helpers are provided separately from the main package to be able to customize
// Regula server's behaviour and create custom binaries.
package cli

import (
	"context"
	stdflag "flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	isatty "github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Config holds the server configuration.
type Config struct {
	Etcd struct {
		Endpoints   []string      `config:"etcd-endpoints"`
		Namespace   string        `config:"etcd-namespace,required"`
		DialTimeout time.Duration `config:"etcd-dial-timemout"`
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
// The args hold the command line arguments, as found in os.Argv.
// It returns flag.ErrHelp if the -help flag is specified on the command line.
func LoadConfig(args []string) (*Config, error) {
	var cfg Config
	flag := stdflag.NewFlagSet("", stdflag.ContinueOnError)
	flag.StringVar(&cfg.Etcd.Namespace, "etcd-namespace", "", "etcd namespace to use")
	flag.StringVar(&cfg.LogLevel, "log-level", zerolog.DebugLevel.String(), "debug level")
	cfg.Etcd.Endpoints = []string{"127.0.0.1:2379"}
	flag.Var(commaSeparatedFlag{&cfg.Etcd.Endpoints}, "etcd-endpoints", "comma separated etcd endpoints")
	flag.StringVar(&cfg.Server.Address, "addr", "0.0.0.0:5331", "server address to listen on")
	flag.DurationVar(&cfg.Server.Timeout, "server-timeout", 5*time.Second, "server timeout (TODO)")
	flag.DurationVar(&cfg.Server.WatchTimeout, "server-watch-timeout", 30*time.Second, "server watch timeout (TODO)")

	err := confita.NewLoader(env.NewBackend()).Load(context.Background(), &cfg)
	if err != nil {
		return nil, err
	}
	if err := flag.Parse(args[1:]); err != nil {
		return nil, err
	}
	if cfg.Etcd.Namespace == "" {
		return nil, fmt.Errorf("etcdnamespace is required (use the -etc-namespace flag to set it)")
	}
	return &cfg, nil
}

type commaSeparatedFlag struct {
	parts *[]string
}

func (f commaSeparatedFlag) Set(s string) error {
	*f.parts = strings.Split(s, ",")
	return nil
}

func (f commaSeparatedFlag) String() string {
	if f.parts == nil {
		// Note: the flag package can make a new zero value
		// which is how it's possible for parts to be nil.
		return ""
	}
	return strings.Join(*f.parts, ",")
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

// Run the http server on the chosen address. The server will shutdown automatically
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
