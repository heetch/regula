FROM golang:1.10-alpine as builder

RUN apk --update add git ca-certificates && rm -rf /var/cache/apk/*

COPY . /go/src/github.com/heetch/regula

WORKDIR /go/src/github.com/heetch/regula

RUN go install -v github.com/heetch/regula/cmd/re-server
RUN go install -v github.com/heetch/regula/cmd/ruleset-loader

FROM alpine:latest

RUN apk --update add curl

COPY --from=builder /go/bin/re-server /re-server
COPY --from=builder /go/bin/ruleset-loader /ruleset-loader

CMD ["/re-server"]
