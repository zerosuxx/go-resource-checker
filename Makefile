default: help

.PHONY: build

repository=`cat go.mod | head -n 1 | cut -c8-`
version=`git describe --tags`

help: ## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' -e 's/:.*#/ #/'

install: ## Install the binary
	go install
	go get golang.org/x/lint/golint

build: ## Build the application
	go build -o build/resource-checker checker.go

build-all: ## Build application for supported architectures
	@echo "repository: ${repository}"
	@echo "version: ${version}"
	GOOS=linux GOARCH=amd64 go build -ldflags="-X '${repository}/main.Version=${version}'" -o build/${BINARY_NAME}-linux-x86_64 checker.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-X '${repository}/main.Version=${version}'" -o build/${BINARY_NAME}-linux-aarch64 checker.go
	GOOS=linux GOARCH=arm go build -ldflags="-X '${repository}/main.Version=${version}'" -o build/${BINARY_NAME}-linux-armv7l checker.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X '${repository}/main.Version=${version}'" -o build/${BINARY_NAME}-darwin-x86_64 checker.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X '${repository}/main.Version=${version}'" -o build/${BINARY_NAME}-darwin-aarch64 checker.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-X '${repository}/main.Version=${version}'" -o build/${BINARY_NAME}-windows-x86_64.exe checker.go

run: ## Run the application
	go run checker.go server

lint: ## Check lint errors
	golint -set_exit_status=1 -min_confidence=1.1 ./...