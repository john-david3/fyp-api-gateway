GOCMD  = go
GORUN  = $(GOCMD) run
GOTEST = $(GOCMD) test

PROJECT_DIR = cmd
MAIN        = main.go

run-api:
	@echo "Running microservices"
	@cd ${PROJECT_DIR} && ${GORUN} ${MAIN}

unit-tests:
	@echo "Starting Unit Tests"
	@${GOTEST} ./...