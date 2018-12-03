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
	regula -log-level debug -etcd-namespace regula-local -server-dist-path ./ui/app/dist

run-ui:
	cd ./ui/app && yarn serve

build-ui:
	cd ./ui/app && yarn build
