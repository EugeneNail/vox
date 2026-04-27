FROM golang:1.26.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/realtime-service ./cmd/app/main.go

CMD ["realtime-service"]
