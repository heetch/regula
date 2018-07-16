NAME := regula

.PHONY: all $(NAME) test testrace

all: $(NAME)

$(NAME):
	go install ./cmd/$@

test:
	go test -v -cover -timeout=1m ./...

testrace:
	go test -v -race -cover -timeout=2m ./...
