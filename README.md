# rules-engine

[![Build Status](https://drone.heetch.net/api/badges/heetch/rules-engine/status.svg)](https://drone.heetch.net/heetch/rules-engine)

## Install

```sh
go get github.com/heetch/rules-engine
```

## Usage

```go
package main

import (
  "log"

  rules "github.com/heetch/rules-engine"
)

func main() {
  store, err := consul.NewStore("127.0.0.1:8500", "/rules")
  if err != nil {
    log.Fatal(err)
  }

  engine := rules.NewEngine(store)
  val, err := engine.GetString("/a/b/c", rule.Params{
    "foo": "bar",
  })
  switch err {
    case rules.ErrRuleNotFound:
      // when the rule doesn't exist
    case rules.ErrTypeMismatch:
      // when the rule returns the bad type
    case nil:
      // everything is fine
    default:
      // something unexpected happened
  }
}
```
