FROM golang:1.26.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/migrator ./cmd/migrator
RUN go build -o /usr/local/bin/profile-migrator ./cmd/migrator/main.go

CMD ["profile-migrator"]
