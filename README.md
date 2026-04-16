# Final Year Project API Gateway
A three plane API Gateway for Microservices. Written primarily in Go with NGINX functioning as the data plane. A management plane written in HTML, CSS and JS offers users a user interface to configure their microservice APIs.

## Requirements
- `Golang v1.24.11`
- `Docker-compose` for self hosting
- `Make` for easy setup

## How to Deploy
1. Clone the Repository
2. With make installed, in the root directory run the following:
```
make docker-build
make docker-run
```
3. If not using a load balancer, the management plane is hosted on port `81`
4. If not using a load balancer, the data plane can be accessed on port `8080`

## How to Stop
Simply run the following in the root directory:
```
make docker-stop
```
