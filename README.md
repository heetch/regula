# rules-engine

[![Build Status](https://drone.heetch.net/api/badges/heetch/rules-engine/status.svg)](https://drone.heetch.net/heetch/rules-engine)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.dev.heetch.internal/pkg/github.com/heetch/rules-engine/)

## Install

```sh
go get github.com/heetch/rules-engine
```

## Documentation

For detailed documentation and basic usage examples, please see the package
documentation at <http://godoc.dev.heetch.internal/pkg/github.com/heetch/rules-engine>.

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

  store, err := etcd.NewStore(cli, "prefix")
  if err != nil {
    log.Fatal(err)
  }

  engine := rules.NewEngine(store)

  val, err := engine.GetString("/a/b/c", rule.Params{
    "paramA": "valA",
    "paramB": "valB",
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
