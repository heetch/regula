# Regula

[![Build Status](https://travis-ci.org/heetch/regula.svg?branch=master)](https://travis-ci.org/heetch/regula)
[![ReadTheDocs](https://readthedocs.org/projects/regula/badge/?version=latest&style=flat)](https://regula.readthedocs.io/en/latest/)
[![GoDoc](https://godoc.org/github.com/heetch/regula?status.svg)](https://godoc.org/github.com/heetch/regula)

Regula is an open source Business Rules Engine solution.

:warning: *Please note that Regula is an experiment and that the API is currently considered unstable.*

## Documentation

Comprehensive documentation is viewable on Read the Docs:

https://regula.readthedocs.io/en/latest/

## Building from source

### Server

Installing dependencies

```sh
dep ensure
```

Building

```sh
make
```

Running the server

```sh
make run
```

### UI

Installing dependencies

```sh
cd ./ui/app && yarn install
```

Running the UI dev-server

```sh
make run-ui
```

### Running the server with the UI

```sh
make build-ui
make run
```
