GOCMD  = go
GORUN  = $(GOCMD) run
GOTEST = $(GOCMD) test

DOCKER      = docker
DOCKERRUN   = $(DOCKER) run
DOCKERBUILD = $(DOCKER) build

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
	${DOCKERBUILD} -f dataplane/Dockerfile -t gateway .
	${DOCKERBUILD} -f services/Dockerfile -t test-services .

docker-run:
	@echo "Running docker containers"
	${DOCKERRUN} -d --name services -p 9001:9001 -p 9002:9002 test-services
	${DOCKERRUN} -d --name gateway -p 8080:8080 --link services gateway

docker-stop:
	@echo "Killing all containers"


test-routes:
	@echo "sending routing requests"
	curl "${HOST}/products"
	curl "${HOST}/orders"