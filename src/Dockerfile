FROM golang:1.24.11-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./src/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
RUN chmod +x /app/main
RUN apk update && apk add --no-cache apache2-utils
RUN mkdir -p /etc/nginx/users

EXPOSE 9090

CMD ["./main"]