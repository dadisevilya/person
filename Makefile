SERVICE_NAME := $(shell basename $$PWD)
PID          = /tmp/${SERVICE_NAME}.pid
BINARY_PATH  = bin/service
GO_FILES     = cmd/backend/*.go
HTTP_PORT := $(if $(HTTP_PORT),$(HTTP_PORT),8080)

run: restart
	fswatch -o cmd pkg config | xargs -n1 -I{} make restart || make kill

kill:
	kill `cat $(PID)` || true

build:
	GO111MODULE=on go build -o $(BINARY_PATH) $(GO_FILES)

restart: kill build
	HTTP_PORT=$(HTTP_PORT) $(BINARY_PATH) & echo $$! > $(PID)

fmt:
	go fmt ./...

test:
	APP_ENV=test APP_CONF_PATH=$(shell pwd)/config go test -v -count=1 ./...

clean:
	rm bin/*

.PHONY: run kill build restart fmt test # let's go to reserve rules names

