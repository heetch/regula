FROM golang:1.10-alpine

RUN apk --update add git ca-certificates && rm -rf /var/cache/apk/*

COPY . /go/src/github.com/heetch/rules-engine

WORKDIR /go/src/github.com/heetch/rules-engine

RUN go install -v github.com/heetch/rules-engine/cmd/re-server

CMD ["/go/bin/re-server"]
