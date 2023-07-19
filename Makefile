all: tidy fmt test build

.PHONY: build

APP_NAME = yae

ifeq ($(GOPATH),)
	PATH := $(HOME)/go/bin:$(PATH)
else
	PATH := $(GOPATH)/bin:$(PATH)
endif

export GO111MODULE=on

build: 
	go build

deps:
	go mod download

test:
	go test -cover ./...

fmt:
	gofumpt -w -l .

tidy:
	go mod tidy
