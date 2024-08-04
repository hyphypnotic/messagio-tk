# Используйте официальный образ Golang в качестве базового образа
FROM golang:1.22.5 AS builder

# Установите рабочий каталог
WORKDIR /app

# Скопируйте go.mod и go.sum и загрузите зависимости
COPY go.mod go.sum ./
RUN go mod download

# Скопируйте исходный код в контейнер
COPY . .

# Соберите ваше приложение для Linux архитектуры
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./cmd/msg_handler/main.go

# Используйте официальный образ Alpine для более легкого контейнера
FROM alpine:latest

# Установите рабочий каталог
WORKDIR /app

# Установите зависимости для работы бинарного файла (опционально, если требуется)
RUN apk add --no-cache ca-certificates

# Скопируйте бинарный файл из этапа сборки
COPY --from=builder /app/main .

# Скопируйте файл конфигурации в контейнер
COPY config/local.yaml /app/config/local.yaml

# Сделайте файл исполняемым
RUN chmod +x ./main


# Запустите приложение
CMD ["./main"]
