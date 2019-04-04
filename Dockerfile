FROM golang:1.12-alpine as builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
RUN apk --no-cache --update add git
WORKDIR $GOPATH/src/github.com/heetch/regula
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /regula ./cmd/regula
RUN chmod +x /regula
CMD ["/regula"]

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /regula .
EXPOSE 5331/tcp
CMD ["./regula"]
