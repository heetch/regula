package cli

import (
	"context"
	stdflag "flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/regula/api/server"
	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

// Config holds the server configuration.
type Config struct {
	Etcd struct {
		Endpoints    []string `config:"etcd-endpoints"`
		DefaultLimit int      `config:"etcd-default-limit"`
		Namespace    string   `config:"etcd-namespace"`
	}
	Server struct {
		Address      string        `config:"addr"`
		Timeout      time.Duration `config:"server-timeout"`
		WatchTimeout time.Duration `config:"server-watch-timeout"`
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
	flag.IntVar(&cfg.Etcd.DefaultLimit, "etcd-default-limit", 50, "etcd default limit for Get requests")
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
