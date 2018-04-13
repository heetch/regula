# rules-engine

[![Build Status](https://drone.heetch.net/api/badges/heetch/rules-engine/status.svg)](https://drone.heetch.net/heetch/rules-engine)
[![Godoc](https://img.shields.io/badge/doc-latest-blue.svg)](http://godoc.dev.heetch.internal/github.com/heetch/rules-engine/)

## Install

```sh
glide get github.com/heetch/rules-engine
```

## Documentation

For detailed documentation and basic usage examples, please see the package
documentation at <http://godoc.dev.heetch.internal/github.com/heetch/rules-engine>.

## Usage

```go
package main

import (
  "log"
  "time"

  "github.com/coreos/etcd/clientv3"
  rules "github.com/heetch/rules-engine"
  "github.com/heetch/rules-engine/etcd"
  "github.com/heetch/rules-engine/rule"
)

func main() {
  cli, err := clientv3.New(clientv3.Config{
    Endpoints:   []string{":2379"},
    DialTimeout: 5 * time.Second,
  })
  if err != nil {
    log.Fatal(err)
  }
  defer cli.Close()

  store, err := etcd.NewStore(cli, etcd.Options{
    Prefix: "prefix",
    Logger: log.New(os.Stdout, "[etcd] ", log.LstdFlags),
  })
  if err != nil {
    log.Fatal(err)
  }
  defer store.Close()

  engine := rules.NewEngine(store)

  val, err := engine.GetString("/a/b/c", rule.Params{
    "product-id": "1234",
    "user-id"   : "5678",
  })
  if err != nil {
    switch err {
    case rules.ErrRulesetNotFound:
      // when the ruleset doesn't exist
    case rules.ErrTypeMismatch:
      // when the ruleset returns the bad type
    case rule.ErrNoMatch:
      // when the ruleset doesn't match
    default:
      // something unexpected happened
    }
  }
}
```

## Tools

The [tools](tools/etcd-ruleset-creator/README.md) directory contains a tool to easily create rulesets on etcd.
