package main

import (
	"log"
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

	logger := cli.CreateLogger(cfg.LogLevel)

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	defer etcdCli.Close()

	service := etcd.RulesetService{
		Client:    etcdCli,
		Namespace: cfg.Etcd.Namespace,
	}

	srv := server.New(&service, server.Config{
		Logger:       &logger,
		WatchTimeout: cfg.Server.WatchTimeout,
	})

	cli.RunServer(srv, cfg.Server.Address)
}
