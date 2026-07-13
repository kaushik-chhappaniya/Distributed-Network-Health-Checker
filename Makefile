# Makefile for docker setup and teardown

NODES =? 1
ENGINE ?= docker
COMPOSE ?= docker-compose
OUTPUT ?= docker-compose.yml

DOCKER_COMPOSE_FILE := ./compose/${OUTPUT}

.DEFAULT_GOAL := help

# Show help message
help:
	@echo "Hello, welcome to docker setup util"
	@echo "============================================"
	@echo "Docker Health Checkup Setup"
	@echo "============================================"
	@echo "Available Targets"
	@echo "make create-docker-compose            Creates docker compose file"
	@echo "make run-docker-compose            	 Run docker compose file"
	@echo "make teardown            			 Teardown all contaiers along with volumes"
	@echo "make help				             Display this menu"
	@echo "make help				             Display this menu"
.PHONY:help

# Create sample docker compose 
create-docker-compose:
	@echo "Creating a sample docker compose file ${DOCKER_COMPOSE_FILE} with ${NODES} nodes"
	@go run ./cmd/composegen/main.go \
		--nodes ${NODES} \
		-output "${DOCKER_COMPOSE_FILE}"
	@if [ ! -f  ${DOCKER_COMPOSE_FILE} ]; then \
		echo "Failed to create docker compose file"; \
		exit 1; \
	fi
	@echo "Docker compose file ${DOCKER_COMPOSE_FILE} created successfully with ${NODES} nodes"
.PHONY: create-docker-compose

# Run docker compose
run-docker-compose:
	@echo "Running docker compose"
	@${COMPOSE} -f ${DOCKER_COMPOSE_FILE} up -d
	@echo "Docker compose is running"
.PHONY: run-docker-compose

# Tear down docker compose along with all volumes
teardown:
	@echo "Tearing down docker compose"
	@${COMPOSE} -f ${DOCKER_COMPOSE_FILE} down -v
	@echo "Docker compose is torn down"
.PHONY: teardown