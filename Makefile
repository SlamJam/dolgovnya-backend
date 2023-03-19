
GOLANGBIN:=$(shell go env GOPATH)/bin
USE_PODMAN?=1

# ===
DOCKER:=$(if $(USE_PODMAN),podman,docker)
COMPOSE:=$(if $(USE_PODMAN),podman-compose,"docker compose")

define assert_nonempty
  $(if $(strip $(1)),,$(error $(2)))
endef

.PHONY: all
all: build

.PHONY: install/tools
install/tools:
	@go install github.com/pressly/goose/v3/cmd/goose@v3
	@go install github.com/bufbuild/buf/cmd/buf@v1.15.1
	@go install github.com/mitchellh/gox@v1.0.1

.PHONY: migration/new
migration/new:
	$(call assert_nonempty, ${name}, "Must set migration name as name=my-super-migration")
	@ ${GOLANGBIN}/goose -dir migrations create "${name}" sql

.PHONY: proto/format
proto/format:
	@ ${GOLANGBIN}/buf format api/ -w


.PHONY: proto/lint
proto/lint:
	@ ${GOLANGBIN}/buf lint api/

.PHONY: proto/generate
proto/generate: proto/lint
	@ find internal/pb -mindepth 1 -delete
	@ find internal/swagger -name '*.swagger.json' -type f -mindepth 1 -maxdepth 1 -delete
	@ ${GOLANGBIN}/buf generate api/

.PHONY: build
build:
	@ go build -o ./bin/dolgovnya .

.PHONY: local/up
local/up:
	@ ${COMPOSE} -f .local/docker-compose.yaml up -d

.PHONY: local/down
local/down:
	@ echo down
	@ ${COMPOSE} -f .local/docker-compose.yaml down

.PHONY: local/prune
local/prune:
	@ echo prune
	@ ${COMPOSE} -f .local/docker-compose.yaml down -v

.PHONY:local/cleanstate
local/cleanstate: local/prune local/up
	@ go run . migration up

.PHONY: docker/build
docker/build:
	@ ${DOCKER} build . -f .build/Dockerfile
