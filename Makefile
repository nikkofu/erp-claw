GOCACHE_DIR := $(CURDIR)/.cache/go-build
DOCKER_ENV_FILE := configs/local/docker.env
DOCKER_COMPOSE := docker compose --env-file $(DOCKER_ENV_FILE)

.PHONY: infra-up infra-down infra-reset api gateway worker scheduler test test-integration smoke migrate-up migrate-down

infra-up:
	$(DOCKER_COMPOSE) up -d --build

infra-down:
	$(DOCKER_COMPOSE) down

infra-reset:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	$(MAKE) infra-up

api:
	go run ./cmd/api-server

gateway:
	go run ./cmd/agent-gateway

worker:
	go run ./cmd/worker

scheduler:
	go run ./cmd/scheduler

test:
	go test ./...

test-integration:
	@mkdir -p $(dir $(GOCACHE_DIR))
	GOCACHE=$(GOCACHE_DIR) go test ./test/integration -run TestDockerComposeIncludesRequiredServices -v

smoke:
	./scripts/smoke_local.sh

migrate-up:
	go run ./cmd/migrate

migrate-down:
	ERP_MIGRATIONS_DIRECTION=down go run ./cmd/migrate
