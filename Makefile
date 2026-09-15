# Makefile for docker compose and containerlab running

DOCKER_COMPOSE_FILE  := ./compose/docker-compose.yml
NETWORK_COMPOSE_FILE := ./compose/docker-network-compose.yml
DOCKERFILE           := ./compose/Dockerfile
DEBUG_DOCKERFILE     := ./compose/Dockerfile.debug

HEALTH_CHECK_CONTAINER := health-checker

.DEFAULT_GOAL := help

# Show help message
help:
	@echo "Hello, welcome to docker setup util"
	@echo "================================================================"
	@echo "Docker Health Checkup Setup"
	@echo "================================================================"
	@echo "Available Targets:"
	@echo "  make build               Build the docker compose images"
	@echo "  make up                  Start all services (db & api) in background"
	@echo "  make up-db               Start only the Postgres database service"
	@echo "  make restart-app         Rebuild & restart ONLY the Go app container"
	@echo "  make run-health-check    Run the health-checker inside api container"
	@echo "  make logs                Tail logs from docker compose services"
	@echo "  make down                Stop and remove containers and volumes"
	@echo "  make run-network         Start Alpine test nodes (docker-network-compose)"
	@echo "  make teardown-network    Stop Alpine test nodes (docker-network-compose)"
	@echo "  make help                Display this menu"
.PHONY: help

# ------------------------------------------------ COMPOSE -------------------------------------------------

# Build the containers
build:
	@echo "Building docker compose services..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} build
	@echo "Build complete."
.PHONY: build

# Run all services in background
up:
	@echo "Starting all services (db, api)..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} up -d
	@echo "Services started."
.PHONY: up

# Run Postgres service only
up-db:
	@echo "Starting Postgres service..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} up -d db
	@echo "Postgres service started."
.PHONY: up-db

# Rebuild and restart only the Go app container (without affecting DB or other nodes)
restart-app:
	@echo "Rebuilding and restarting Go api container..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} up -d --build api
	@echo "Go api container updated."
.PHONY: restart-app

# Run health check binary inside running api container
run-health-check:
	@echo "Executing health-checker inside api container..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} exec api ./health-checker
.PHONY: run-health-check

# Tail logs of docker compose services
logs:
	@docker compose -f ${DOCKER_COMPOSE_FILE} logs -f
.PHONY: logs

# Teardown all compose containers and volumes
down:
	@echo "Tearing down docker compose containers and volumes..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} down -v
	@echo "Teardown complete."
.PHONY: down

# ------------------------------------------------ NETWORK NODES -------------------------------------------------

# Spin up alpine network nodes
run-network:
	@echo "Starting Alpine network nodes..."
	@docker compose -f ${NETWORK_COMPOSE_FILE} up -d
	@echo "Alpine nodes are running on health-net."
.PHONY: run-network

# Teardown alpine network nodes
teardown-network:
	@echo "Stopping Alpine network nodes..."
	@docker compose -f ${NETWORK_COMPOSE_FILE} down -v
	@echo "Alpine network nodes stopped."
.PHONY: teardown-network