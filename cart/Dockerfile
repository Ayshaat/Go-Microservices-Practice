FROM golang:1.24-alpine

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY . .

RUN go build -mod=vendor -o cart ./cmd/main.go

ENV HTTP_PORT=8090

EXPOSE 8090

CMD ["./cart"]
