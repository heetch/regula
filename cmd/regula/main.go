package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula/api/server"
	"github.com/heetch/regula/cmd/regula/cli"
	"github.com/heetch/regula/store/etcd"
)

func main() {
	cfg, err := cli.LoadConfig(os.Args)
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "regula: %v\n", err)
		os.Exit(2)
	}

	logger := cli.CreateLogger(cfg.LogLevel, os.Stderr)

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to etcd cluster")
	}
	defer etcdCli.Close()

	service := etcd.RulesetService{
		Client:    etcdCli,
		Namespace: cfg.Etcd.Namespace,
		Logger:    logger.With().Str("service", "etcd").Logger(),
	}

	srv := server.New(&service, server.Config{
		Logger:       &logger,
		Timeout:      cfg.Server.Timeout,
		WatchTimeout: cfg.Server.WatchTimeout,
	})

	cli.RunServer(srv, cfg.Server.Address)
}
