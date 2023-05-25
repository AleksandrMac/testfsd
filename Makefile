# OS
OSNAME						:=
BINARY_NAME_FILE	:=
ifeq ($(OS),Windows_NT)
	OSNAME=windows
else
	UNAME_S :=$(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OSNAME=linux
	endif
	ifeq ($(UNAME_S),Darwin)
		OSNAME=drawin
	endif
endif

# Env
CGO_ENABLED=1
GOCMD=go
GOARCH=amd64
BINARY_NAME=server
BINARY_NAME_FILE =./dist/$(OSNAME)/
BINARY_NAME_LINUX=./dist/linux/
BINARY_NAME_MACOS=./dist/drawin/
BINARY_NAME_WIN=./dist/windows/
GIT_COMMIT=$(shell git rev-list -1 HEAD)
VERSION=$(shell date "+%Y.%m.%d.%H:%M:%S")
BUILD_DATE=$(shell date "+%Y.%m.%d.%H:%M:%S")
GIT_TAG=$(shell git describe --tags)
BUILD_FLAGS="-X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(GIT_TAG) -X main.BuildDate=$(BUILD_DATE)"

PROJECT_NAME = kpi_tinkoff
DOCKER_REGISTRY_NAME = cr.yandex/crpvule9pdaospoejrev
DOCKER_IMAGE_NAME = kpi-tinkoff
prebuild:
	mkdir -p ./dist/$(OSNAME)/
prebuild-all:
	mkdir -p $(BINARY_NAME_LINUX)
	mkdir -p $(BINARY_NAME_MACOS)
	mkdir -p $(BINARY_NAME_WIN)
_dist_os:
	$(GOCMD) build $(BUILD_FLAGS) -o $(BINARY_NAME_FILE) ./cmd/...
build: prebuild _dist_os
build-linux:
	CC="x86_64-linux-musl-gcc" CXX="x86_64-linux-musl-g++" GOOS=linux $(GOCMD) build -v -ldflags $(BUILD_FLAGS) -o $(BINARY_NAME_LINUX) ./cmd/...
build-mac:
	GOOS=darwin $(GOCMD) build $(BUILD_FLAGS) -o $(BINARY_NAME_MACOS) ./cmd/...
build-win:
	CC="x86_64-w64-mingw32-gcc" GOOS=windows $(GOCMD) build $(BUILD_FLAGS) -o $(BINARY_NAME_WIN) ./cmd/...
test:
	$(GOCMD) test -v ./...
clean:
	$(GOCMD) clean ./...
	rm -rf ./dist/
doc:
	$(GOCMD) run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/start.go --parseDependency
download:
	go mod download
build-all: build-mac build-win build-linux
all: test prebuild-all build-all
run:
	$(GOCMD) run ./cmd/server start
build-image:
	docker-compose build
run-docker:
	docker-compose up --build
docker_build:
	docker build -f build/docker/Dockerfile  \
  --tag ${DOCKER_REGISTRY_NAME}/${DOCKER_IMAGE_NAME}:latest \
  --tag ${DOCKER_REGISTRY_NAME}/${DOCKER_IMAGE_NAME}:$(GIT_TAG) \
  --build-arg BUILD_FLAGS=$(BUILD_FLAGS) .
docker_push:
	docker push --all-tags ${DOCKER_REGISTRY_HOST}${DOCKER_REGISTRY_NAME}/${DOCKER_IMAGE_NAME}