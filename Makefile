GOCMD  = go
GORUN  = $(GOCMD) run
GOTEST = $(GOCMD) test

DOCKER      = docker
COMPOSE     = $(DOCKER) compose
PROJECT_DIR = cmd
MAIN        = main.go

HOST = http://localhost:8080

WAIT_RETRIES=5
WAIT_INTERVAL=3

run:
	@echo "Running microservices"
	@cd ${PROJECT_DIR} && ${GORUN} ${MAIN}

unit-tests:
	@echo "Starting Unit Tests"
	@${GOTEST} ./...

docker-build:
	@echo "Building Docker Files"
	@${COMPOSE} build

docker-run:
	@echo "Running docker containers"
	@${COMPOSE} up -d

docker-stop:
	@echo "Killing all containers"
	@${COMPOSE} down

wait:
	@echo "Waiting for $(SERVICE) on port $(PORT)..."
	@for i in $(shell seq 1 $(WAIT_RETRIES)); do \
		if curl -s http://localhost:$(PORT)/healthz >/dev/null 2>&1; then \
			echo "$(SERVICE) is ready"; \
			break; \
		fi; \
		echo "Waiting... ($$i)"; \
		sleep $(WAIT_INTERVAL); \
	done

test-docker: docker-build docker-run wait
	@$(MAKE) docker-stop

test-routes:
	@echo "sending routing requests"
	curl "${HOST}/products"
	curl "${HOST}/orders"
	curl "${HOST}/protected"