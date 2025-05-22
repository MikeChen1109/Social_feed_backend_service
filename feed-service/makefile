APP_NAME := main
GO_FILES := $(shell find . -name '*.go' -type f)

.PHONY: run test all

run:
	go run -race $(APP_NAME).go

test:
	go test -race ./... -v -coverprofile=./cover.out

all: run test