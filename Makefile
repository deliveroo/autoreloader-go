APP := $(shell basename $$(pwd))

.PHONY: build
build:
	go build -o build/$(APP) main.go

.PHONY: install
install:
	go install -v ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: docker-test
docker-test:
	@docker build -t $(APP):latest -f Dockerfile .
