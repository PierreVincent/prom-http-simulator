# Multi-stage Dockerfile

FROM golang:1.9
WORKDIR /go/src/github.com/PierreVincent/prom-http-simulator/
COPY vendor vendor
COPY cmd cmd
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -v -o bin/prom-http-simulator cmd/main.go

FROM alpine:3.6
COPY --from=0 /go/src/github.com/PierreVincent/prom-http-simulator/bin/prom-http-simulator /usr/local/bin/prom-http-simulator
ENTRYPOINT ["prom-http-simulator"]