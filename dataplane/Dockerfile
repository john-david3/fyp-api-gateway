FROM golang:1.24.11-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./dataplane/main.go

FROM nginx:alpine

ARG HTPASSWD_USER
ARG HTPASSWD_PASS

RUN rm -f /etc/nginx/nginx.conf

COPY --from=builder /app/main .
COPY dataplane/nginx/nginx.conf /etc/nginx/nginx.conf

RUN mkdir -p /etc/nginx/users
RUN apk update && apk add --no-cache apache2-utils
RUN mkdir -p /etc/apache2 && \
    htpasswd -cb /etc/apache2/.htpasswd "$HTPASSWD_USER" "$HTPASSWD_PASS"


EXPOSE 8080

CMD ["sh", "-c", "./main & exec nginx -g 'daemon off;'"]
