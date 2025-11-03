FROM golang:1.24-alpine

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY *.go ./

# Собираем приложение
RUN go build -o /bionicpro-auth

EXPOSE 8000

CMD ["/bionicpro-auth"]