NAME := regula

.PHONY: all $(NAME) test testrace run build run-ui build-ui

all: $(NAME)

build: $(NAME)

$(NAME):
	go install ./cmd/$@

test:
	go test -v -cover -timeout=1m ./...

testrace:
	go test -v -race -cover -timeout=2m ./...

run: build
	LOG_LEVEL=debug regula -etcd-namespace regula-local

run-ui:
	cd ./ui/app && yarn serve

build-ui:
	cd ./ui/app && yarn build
