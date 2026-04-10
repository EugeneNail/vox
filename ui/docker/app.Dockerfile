FROM node:24-alpine AS dev

WORKDIR /app/internal

COPY internal/package.json internal/package-lock.json ./
RUN npm ci

COPY internal ./

FROM node:24-alpine AS frontend-builder

WORKDIR /app/internal

COPY internal/package.json internal/package-lock.json ./
RUN npm ci

COPY internal ./
RUN npm run build

FROM golang:1.26.1 AS go-builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY 'cmd' './cmd'
COPY internal/frontenddist.go ./internal/frontenddist.go
COPY --from=frontend-builder /app/internal/dist ./internal/dist
RUN go build -o /usr/local/bin/ui-service ./cmd/app/main.go

FROM debian:trixie-slim AS prod

COPY --from=go-builder /usr/local/bin/ui-service /usr/local/bin/ui-service

CMD ["/usr/local/bin/ui-service"]
