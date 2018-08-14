package main

import (
	"log"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula/api/server"
	"github.com/heetch/regula/cmd/regula/cli"
	"github.com/heetch/regula/store/etcd"
)

func main() {
	cfg, err := cli.LoadConfig()
	if err != nil {
		log.Fatal(err)
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
	}

	srv := server.New(&service, server.Config{
		Logger:       &logger,
		Timeout:      cfg.Server.Timeout,
		WatchTimeout: cfg.Server.WatchTimeout,
	})

	cli.RunServer(srv, cfg.Server.Address)
}
