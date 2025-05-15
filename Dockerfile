# Вибираємо офіційний образ Golang як builder
FROM golang:1.24 AS builder

WORKDIR /app

# Копіюємо файли модуля і завантажуємо залежності
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код
COPY . .

# Компільовуємо бінарник у папку /app/bin
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/weather_subscriber ./main.go

# Другий етап: мінімальний образ для запуску
FROM alpine:3.18

# Копіюємо бінарник з першого етапу
COPY --from=builder /app/bin/weather_subscriber /usr/local/bin/weather_subscriber

# Копіюємо конфігурацію, якщо є (опціонально)
# COPY config.yaml /etc/weather_subscriber/config.yaml

# Виставляємо робочу папку
WORKDIR /app

# Відкриваємо порт
EXPOSE 8080

# Запускаємо додаток
CMD ["weather_subscriber"]
