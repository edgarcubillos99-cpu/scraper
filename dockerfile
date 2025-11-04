# Build stage
FROM golang:1.21-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /scraper ./cmd/scraper

# Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=build /scraper /root/scraper
# opcional: archivo .env será montado por docker-compose
CMD ["/root/scraper"]
