package main

import (
	"log"

	"github.com/heetch/regula/cmd/regula/cli"
	"github.com/heetch/regula/store/etcd"
)

func main() {
	cfg, err := cli.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := cli.CreateLogger(cfg.LogLevel)

	etcdCli, err := cli.EtcdClient(cfg.Etcd.Endpoints)
	if err != nil {
		log.Fatal(err)
	}
	defer etcdCli.Close()

	service := etcd.RulesetService{
		Client:    etcdCli,
		Namespace: cfg.Etcd.Namespace,
	}

	srv := cli.CreateServer(cfg, &service, logger)

	cli.RunServer(srv, cfg.Server.Address)
}
