FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o metrics-consumer ./cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/metrics-consumer .
<<<<<<< HEAD
COPY .env.docker .env
=======
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318

ENTRYPOINT ["./metrics-consumer"]
