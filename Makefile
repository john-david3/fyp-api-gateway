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
	${DOCKERBUILD} --progress=plain --no-cache -f dataplane/Dockerfile -t data-plane .
	${DOCKERBUILD} --no-cache -f services/Dockerfile -t test-services .
	${DOCKERBUILD} --no-cache --progress=plain -f src/Dockerfile -t control-plane .

docker-run:
	@echo "Running docker containers"
	docker network create gateway-net || true
	${DOCKERRUN} -d --name services --network gateway-net -p 9001:9001 -p 9002:9002 test-services
	${DOCKERRUN} -d --name control-plane --network gateway-net -p 8081:8081 \
			-v ./config:/etc/config \
 			-v ./src/templates/nginx.conf.tmpl:/etc/nginx/nginx.conf.tmpl:ro \
			control-plane
	${DOCKERRUN} -d --name data-plane --network gateway-net -p 8080:8080 data-plane

docker-stop:
	@echo "Killing all containers"
	${DOCKERKILL} services
	${DOCKERKILL} data-plane
	${DOCKERKILL} control-plane

test-routes:
	@echo "sending routing requests"
	curl "${HOST}/products"
	curl "${HOST}/orders"
	curl "${HOST}/test"