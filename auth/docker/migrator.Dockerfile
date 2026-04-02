FROM golang:1.26.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/migrator ./cmd/migrator
RUN go build -o /usr/local/bin/auth-migrator ./cmd/migrator/main.go

CMD ["auth-migrator"]
