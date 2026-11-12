# Makefile for docker compose and containerlab running

DOCKER_COMPOSE_FILE  := ./compose/docker-compose.yml
NETWORK_COMPOSE_FILE := ./compose/docker-network-compose.yml
DOCKERFILE           := ./compose/Dockerfile
DEBUG_DOCKERFILE     := ./compose/Dockerfile.debug
LAB_FILE 			 := ./compose/topology/topology.clab.yml
LAB_GEN_FILES		 := ./compose/topology/clab*

HEALTH_CHECK_CONTAINER := health-checker

.DEFAULT_GOAL := help

# Show help message
help:
	@echo "Hello, welcome to docker setup util"
	@echo "================================================================"
	@echo "Docker Health Checkup Setup"
	@echo "================================================================"
	@echo "Available Targets:"
	@echo "Full Stack:"
	@echo "  make up                  Deploy the network lab and app"
	@echo "  make down                Teardown the network lab and app"
	@echo ""
	@echo "  make app-up              Start all services (db & api) in background"
	@echo "  make db-up               Start only the Postgres database service"
	@echo "  make rebuild-app         Rebuild and restart ONLY the Go API container"
	@echo "  make run-cli             Run the health-checker CLI inside api container"
	@echo "  make psql                Open PSQL interactive CLI"
	@echo "  make logs                Tail logs from all docker compose services"
	@echo "  make logs-api            Tail logs ONLY from the Go API service"
	@echo "  make api-sh              Open interactive shell inside running api container"
	@echo "  make app-down            Stop and remove containers and volumes"
	@echo "  make run-network         Start Alpine test nodes (docker-network-compose)"
	@echo "  make teardown-network    Stop Alpine test nodes (docker-network-compose)"
	@echo "  make lab-up              Deploy Containerlab topology nodes"
	@echo "  make lab-down            Destroy Containerlab nodes"
	@echo "  make help                Display this menu"
.PHONY: help

# ------------------------------------------------------------------------------
# APPLICATION & DATABASE (docker-compose)
# ------------------------------------------------------------------------------

# Start DB and build/run latest Go API Code
app-up:
	@echo "Building and starting DB & API services..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} up -d --build
	@echo "All services running on http://localhost:8080"
	@echo "Services started."
.PHONY: app-up

# Stop all compose services and clean up volumes
app-down:
	@echo "Tearing down docker compose containers and volumes..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} down -v
	@echo "Teardown complete."
.PHONY: app-down

# Run Postgres service only
db-up:
	@echo "Starting Postgres service..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} up -d db
	@echo "Postgres service started."
.PHONY: db-up

# Rebuild and restart only the Go app container (without affecting DB or other nodes)
rebuild-app:
	@echo "Rebuilding and restarting Go api container..."
	@docker compose -f ${DOCKER_COMPOSE_FILE} up -d --build api
	@echo "Go api container updated."
.PHONY: rebuild-app

# ------------------------------------------------------------------------------
# LOGGING
# ------------------------------------------------------------------------------

# Open an interactive shell inside the running API container
api-sh:
	@docker compose -f ${DOCKER_COMPOSE_FILE} exec api sh
.PHONY: api-sh

# Tail logs of docker compose services
logs:
	@docker compose -f ${DOCKER_COMPOSE_FILE} logs -f
.PHONY: logs

# Tail logs of API service only
logs-api:
	@docker compose -f ${DOCKER_COMPOSE_FILE} logs -f api
.PHONY: logs-api

# Shell directly into Postgres CLI
psql:
	@docker compose -f ${DOCKER_COMPOSE_FILE} exec db psql -U admin -d mydatabase
.PHONY: psql

# Run the CLI health-checker tool inside the container (optional)
run-cli:
	@docker compose -f ${DOCKER_COMPOSE_FILE} exec api ./health-checker
.PHONY: run-cli

# ------------------------------------------------------------------------------
# NETWORK LAB (Containerlab / Alpine Nodes)
# ------------------------------------------------------------------------------

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

# Spin up containerlab
lab-up:
	@echo "Deploying Containerlab nodes..."
	sudo containerlab deploy -t $(LAB_FILE)
.PHONY: lab-up

# Teardown containerlab
lab-down:
	@echo "Destroying Container lab nodes..."
	sudo containerlab destroy -t $(LAB_FILE)
	rm -rf $(LAB_GEN_FILES)
.PHONY: lab-down

# ------------------------------------------------------------------------------
# FULL STACK COMPOSITE TARGETS
# ------------------------------------------------------------------------------
up: lab-up app-up
down: app-down lab-down
.PHONY: up down