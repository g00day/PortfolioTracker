FROM golang:1.26-alpine

RUN apk add --no-cache ca-certificates tzdata gcc musl-dev git

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

# Запускаем приложение
CMD ["go", "run", "main.go"]
