# Go parameters
#!make

.DEFAULT_GOAL=test
GOCMD=go
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
include .env
export $(shell sed 's/=.*//' .env)

test:
	$(GOTEST) -v ./...
run:
	$(GORUN) main.go
