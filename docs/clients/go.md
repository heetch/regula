# Golang library quickstart

The Regula Golang library provides programmatic access to the full API and features of Regula.

## Install

```
go get -u github.com/heetch/regula
```

## API documentation

The API documentation can be found on [godoc](https://godoc.org/github.com/heetch/regula).

## Engine

The **engine** is a unified API which allows the evaluation of rulesets.  Whether a ruleset is stored on a remote Regula cluster, or in memory, the evaluation API remains the same. The **engine** delegates the evaluation to an **evaluator** which as its name suggests is able to evaluate rulesets. The role of the **engine** is to make sure the **result** of the evaluation corresponds to the expected type, decodes it and return it to the caller.
This separation allows the decoupling of the evaluation logic from the exploitation of its result by the user.

### Creating an evaluator and an engine

Before being able to use an engine, an evaluator must be instantiated.
Regula provides three ways to do so.

#### Server-side evaluation

One of the simplest methods is to deleguate the evaluation of rulesets to the Regula server.
In order to do that, you must use the Regula Client Ruleset API which implements the `Evaluator` interface.

```go
package main

import (
	"log"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
)

func main() {
	// Create a client.
	cli, err := client.New("http://localhost:5331/")
	if err != nil {
		log.Fatal(err)
	}

	// Create an engine and pass the client.Rulesets field whitch instantiates the regula.Evaluator interface.
	ng := regula.NewEngine(cli.Rulesets)

	// Every call to the engine methods will send a request to the server.
	str, res, err := ng.GetString(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})
}
```

With this evaluator, every call to `GetString` and other methods of the engine object will result in a call to the Regula server.

#### Client side evaluation

Regula also provides client side evaluation to avoid network round-trips when necessary.
At startup, the evaluator loads all the requested rulesets and saves them in a local cache.
An optional mechanism watches the server for changes and automatically updates the local cache.

```go
package main

import (
	"context"
	"log"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
)

func main() {
	// Create a client.
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

	// Create the engine.
	ng := regula.NewEngine(ev)

	// Every call to the engine methods will evaluate rulesets in memory with no network round trip.
	str, res, err := ng.GetString(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})
}
```
