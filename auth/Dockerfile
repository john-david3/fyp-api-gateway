FROM golang:1.24.11-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./auth/main.go


FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
RUN chmod +x /app/main

ENV PRIVATE_KEY_PATH=/etc/keys/private.pem
EXPOSE 8089

CMD ["./main"]
