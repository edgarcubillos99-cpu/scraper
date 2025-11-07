# ========================
# BUILD STAGE
# ========================
FROM golang:1.25.3-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /scraper ./cmd/scraper



# ========================
# FINAL IMAGE
# ========================
FROM alpine:3.18

# ✅ Descargar paquetes necesarios
RUN apk add --no-cache ca-certificates tzdata bash curl mysql-client

WORKDIR /root/

# ✅ Instalar Goose en la imagen final (IMPORTANTE)
RUN curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | bash

# ✅ Copiar binario compilado
COPY --from=build /scraper /root/scraper

# ✅ Copiar entrypoint.sh
COPY entrypoint.sh /root/entrypoint.sh
RUN chmod +x /root/entrypoint.sh

CMD ["/root/entrypoint.sh"]
