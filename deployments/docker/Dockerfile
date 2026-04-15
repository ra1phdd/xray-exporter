FROM golang:1.26-alpine AS build

WORKDIR /build

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOMAXPROCS=1
ARG GOCACHE=/root/.cache/go-build
ARG GOMODCACHE=/go/pkg/mod

ENV CGO_ENABLED=$CGO_ENABLED \
    GOOS=$GOOS \
    GOARCH=$GOARCH \
    GOMAXPROCS=$GOMAXPROCS \
    GOCACHE=$GOCACHE \
    GOMODCACHE=$GOMODCACHE

RUN apk add --no-cache git ca-certificates tzdata

COPY ../../go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY ../.. .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -ldflags="-s -w" \
    -tags netgo,osusergo \
    -o /app/main \
    ./cmd/main/main.go

FROM scratch

WORKDIR /app

EXPOSE 9550

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build --chown=65534:65534 /app/main /main

ENV TZ=Europe/Moscow
USER 65534:65534

ENTRYPOINT ["/main"]
