FROM golang:1.24-bullseye AS builder

RUN apt-get update && apt-get install -y git  

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/app

FROM alpine:3.20

WORKDIR /root/

# Копируем бинарь из builder
COPY --from=builder /app/server .
COPY --from=builder /app/.env .

# Указываем порт
EXPOSE 8080

# Запускаем сервис
CMD ["./server"]
