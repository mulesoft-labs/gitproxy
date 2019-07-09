# Go parameters
#!make

.DEFAULT_GOAL=test
GOCMD=go
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
GODEBUG=dlv debug
include .env
export $(shell sed 's/=.*//' .env)

test:
	$(GOTEST) -v ./...
run:
	$(GORUN) main.go
debug:
	$(GODEBUG) --listen=localhost:54634 --headless=true --api-version=2 main.go
