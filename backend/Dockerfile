FROM golang:1.26-alpine

WORKDIR /app

ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN for attempt in 1 2 3; do \
      go mod download && break; \
      if [ "$attempt" -eq 3 ]; then exit 1; fi; \
      echo "dependency download failed, retrying ($attempt/3)"; \
      sleep 2; \
    done

COPY . .

RUN CGO_ENABLED=0 go build -o /usr/local/bin/portfolio-tracker .

EXPOSE 8080

# Точка входа находится в корне модуля.
CMD ["portfolio-tracker"]
