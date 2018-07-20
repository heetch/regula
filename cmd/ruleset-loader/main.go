// The ruleset-loader reads a json file containing a list of rulesets and sends them to etcd.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
)

var (
	file = flag.String("f", "", "path to file containing rulesets informations")
	addr = flag.String("addr", "127.0.0.1:5331", "regula server address")
)

func main() {
	flag.Parse()

	if *file == "" {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	client, err := client.New(*addr)
	if err != nil {
		log.Fatal(err)
	}

	loadSnapshot(client, f)
}

func loadSnapshot(client *client.Client, r io.Reader) error {
	var snapshot map[string]*regula.Ruleset

	err := json.NewDecoder(r).Decode(&snapshot)
	if err != nil {
		return err
	}

	for path, rs := range snapshot {
		path = strings.TrimSpace(path)
		if path == "" {
			return errors.New("empty path")
		}

		path = strings.Trim(path, "/")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.Rulesets.Put(ctx, path, rs)
		if err != nil {
			return err
		}

		fmt.Printf("Ruleset \"%s\" successfully saved.\n", path)
	}

	return nil
}
