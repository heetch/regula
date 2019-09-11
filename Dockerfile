FROM golang:1.13-alpine as builder

WORKDIR /src/regula
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /regula ./cmd/regula
RUN chmod +x /regula
CMD ["/regula"]

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /regula .
EXPOSE 5331/tcp
CMD ["./regula"]
