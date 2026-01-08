FROM golang:1.24.11-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./dataplane/main.go

FROM nginx:alpine

RUN rm -f /etc/nginx/nginx.conf

COPY --from=builder /app/main .
COPY dataplane/nginx/nginx.conf /etc/nginx/nginx.conf

#RUN chmod +x /app/main

EXPOSE 8080

CMD ["sh", "-c", "./main & exec nginx -g 'daemon off;'"]
