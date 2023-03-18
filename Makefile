
GOLANGBIN:=$(shell go env GOPATH)/bin
USE_PODMAN?=1

# ===
DOCKER:=$(if $(USE_PODMAN),podman,docker)
COMPOSE:=$(if $(USE_PODMAN),podman-compose,"docker compose")

define assert_nonempty
	$(if $(strip $(1)),,$(error $(2)))
endef

.PHONY: install\:\:tools
install\:\:tools:
	@go install github.com/pressly/goose/v3/cmd/goose@v3
	@go install github.com/bufbuild/buf/cmd/buf@v1.15.1
	@go install github.com/mitchellh/gox@v1.0.1

	# @go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.11.2
	# @go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.11.2
	# @go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	# @go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	# @go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@v0.3.0


.PHONY: migration\:\:new
migration\:\:new:
	$(call assert_nonempty, ${name}, "Must set migration name as name=my-super-migration")
	@ ${GOLANGBIN}/goose -dir migrations create "${name}" sql

.PHONY: generate\:\:proto
generate\:\:proto:
	@ find internal/pb -mindepth 1 -delete
	@ find internal/swagger -name '*.swagger.json' -type f -mindepth 1 -maxdepth 1 -delete
	@ ${GOLANGBIN}/buf generate api/

.PHONY: build
build:
	@ go build -o ./bin/dolgovnya .

.PHONY: local\:\:up
local\:\:up:
	@ ${COMPOSE} -f .local/docker-compose.yaml up -d

.PHONY: local\:\:down
local\:\:down:
	@ ${COMPOSE} -f .local/docker-compose.yaml down

.PHONY: local\:\:prune
local\:\:prune:
	@ ${COMPOSE} -f .local/docker-compose.yaml down -v

.PHONY: docker\:\:build
docker\:\:build:
	@ ${DOCKER} build . -f .build/Dockerfile
