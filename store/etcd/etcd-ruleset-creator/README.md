# etcd-ruleset-creator

This tool creates a ruleset and sends it to etcd under the given key and namespace.

## Install dependencies

```sh
glide install
```

## Edit

Open the `main.go` file and edit the `getRuleset` function to create the ruleset of your choosing.

## Build

```sh
go build -o erc
```

## Usage

```sh
$ ./erc -h
Usage of ./erc:
  -addr string
      etcd addr
  -name string
      name of the ruleset
  -namespace string
      prefix to use for namespacing
```

To create a ruleset named `matching/gms/alpha` under the `transportation` namespace:

```sh
$ ./erc -addr 127.0.0.1:2379 -namespace transportation -name matching/gms/alpha
Ruleset "matching/gms/alpha" successfully saved.
```

## Environment

To save rulesets on different environments, use the appropriate VPN and change the `addr` option accordingly.
