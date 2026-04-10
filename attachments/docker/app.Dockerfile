FROM golang:1.26.1

WORKDIR /app

COPY . .
RUN go build -o /usr/local/bin/attachments-service ./cmd/app/main.go

CMD ["attachments-service"]
