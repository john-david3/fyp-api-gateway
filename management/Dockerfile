FROM golang:1.24.11-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/main ./management/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/management/static ./static

EXPOSE 80
RUN chmod +x /app/main
CMD ["./main"]
