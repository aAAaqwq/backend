FROM golang:1.24.9-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 12000

CMD ["./main", "-c", "/app/config/prod.yaml"]