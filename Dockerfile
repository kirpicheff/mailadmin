# === Шаг 1: Сборка Фронтенда ===
FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# === Шаг 2: Сборка Бэкенда ===
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o mailadmin ./main.go

# === Шаг 3: Итоговый Docker-образ ===
FROM alpine:3.18
RUN apk add --no-cache nginx

# Настройка Nginx и статики
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY --from=frontend-builder /app/dist /var/www/mailadmin

# Настройка Бэкенда
COPY --from=backend-builder /app/mailadmin /usr/local/bin/mailadmin

# Скрипт запуска
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Переменные окружения по умолчанию
ENV DB_DSN="postfix:password@tcp(127.0.0.1:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
ENV JWT_SECRET="change-this-secret-key-at-production"
ENV LISTEN_ADDR=":8080"
ENV CORS_ORIGIN="*"

# Проброс веб-порта
EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
