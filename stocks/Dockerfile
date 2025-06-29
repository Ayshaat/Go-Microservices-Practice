FROM golang:1.24-alpine

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY . .

RUN go build -mod=vendor -o stocks ./cmd/main.go

ENV HTTP_PORT=8080

EXPOSE 8080

CMD ["./stocks"]
