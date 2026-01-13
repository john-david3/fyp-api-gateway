GOCMD  = go
GORUN  = $(GOCMD) run
GOTEST = $(GOCMD) test

DOCKER      = docker
DOCKERRUN   = $(DOCKER) run
DOCKERBUILD = $(DOCKER) build
DOCKERKILL  = $(DOCKER) rm -f

PROJECT_DIR = cmd
MAIN        = main.go

HOST = http://localhost:8080

run:
	@echo "Running microservices"
	@cd ${PROJECT_DIR} && ${GORUN} ${MAIN}

unit-tests:
	@echo "Starting Unit Tests"
	@${GOTEST} ./...

docker-build:
	@echo "Building Docker Files"
	@docker compose up --build

docker-run:
	@echo "Running docker containers"
	@docker compose up

docker-stop:
	@echo "Killing all containers"
	@docker compose down

test-routes:
	@echo "sending routing requests"
	curl "${HOST}/products"
	curl "${HOST}/orders"
	curl "${HOST}/test"