# Golang library quickstart

The Regula Golang library provides programmatic access to the full Regula APIs and features.

## Install

```
go get -u github.com/heetch/regula
```

## API documentation

The API documentation can be found on [godoc](https://godoc.org/github.com/heetch/regula).

## Engine

An **engine** is an unified API allowing to evaluate rulesets regardless of their location. Whether a ruleset is stored on a remote Regula cluster or in memory, the evaluation API remains the same. The **engine** deleguates the evaluation to an **evaluator** which as its name suggests is able to evaluate rulesets. The role of the **engine** is to make sure the **result** of the evaluation corresponds to the expected type, decodes it and return it to the caller.
This separation allows to decouple the evaluation logic from the exploitation of its result by the user.

### Creating an evaluator and an engine

Before being able to use an engine, an evaluator must be instantiated.
Regula provides three ways to do so.

#### Server-side evaluation

One of the simplest method is to deleguate the evaluation of rulesets to the Regula server.
In order to do that, the user must create use the Regula Client Ruleset API which implements the `Evaluator` interface.

```go
package main

import (
	"log"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
)

func main() {
	cli, err := client.New("http://localhost:5331/")
	if err != nil {
		log.Fatal(err)
	}

	ng := regula.NewEngine(cli.Rulesets)

	str, res, err := ng.GetString(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})
}
```

With this evaluator, every call to `GetString` and other methods of the engine object will occasionate a call to the Regula server.

#### Client side evaluation

TODO

```go
package main

import (
	"context"
	"log"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
)

func main() {
	cli, err := client.New("http://localhost:5331/")
	if err != nil {
		log.Fatal(err)
	}

	// create a cancelable context to cancel the watching.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Fetch all the rulesets that start with the given prefix and store
	// them in a local cache.
	// The last parameter tells the evaluator to watch the server for changes
	// and to update the local cache.
	// If not necessary (or on mobile), set this to false.
	ev, err := client.NewEvaluator(ctx, cli, "prefix", true)
	if err != nil {
		log.Fatal(err)
	}

	ng := regula.NewEngine(ev)

	str, res, err := ng.GetString(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})
}
```
