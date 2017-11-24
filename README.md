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

## Creating Rules and Rulesets

```go
// if paramA == "a value" == paramB -> "matched A"
rA := rule.New(
  rule.Eq(
    rule.ParamStr("paramA"),
    rule.ValueStr("a value"),
    rule.ParamStr("paramB"),
  ),
  rule.ReturnsStr("matched A"),
)

// if paramA in ["a", "b", "c", otherParam] -> "matched B"
rB := rule.New(
  rule.In(
    rule.ParamStr("paramA"),
    rule.ValueStr("a"),
    rule.ValueStr("b"),
    rule.ValueStr("c"),
    rule.ParamStr("otherParam"),
  ),
  rule.ReturnsStr("matched B"),
)

// default rule -> "matched default"
rDef := rule.New(
  rule.True(), // always match
  rule.ReturnsStr("matched default"),
)

// A ruleset is a list of rules
rs := rule.Ruleset{rA, rB, rDefault}

// Marshal then save to Consul
raw, _ := json.Marshal(rs)
```
