SHELL := /bin/bash

# Include .env if present
-include .env

OS := $(shell uname -s)

check-os:
	@if [ "$(OS)" != "Darwin" ] && [ "$(OS)" != "Linux" ]; then \
	  echo "ERROR: This Makefile only supports macOS or Linux. Found '$(OS)'."; \
	  exit 1; \
	fi

check-docker:
	@which docker > /dev/null 2>&1 || ( echo "ERROR: Docker is not installed. Please install Docker." && exit 1 )
	@docker info > /dev/null 2>&1 || ( echo "ERROR: Docker daemon is not running. Please start Docker." && exit 1 )

# Setup command:
#   - Checks OS
#   - Checks Docker + daemon
#   - Installs tmux if missing
setup: check-os check-docker
	@if [ "$(OS)" = "Darwin" ]; then \
		echo "Running setup on macOS. Checking for tmux..."; \
		brew list tmux >/dev/null 2>&1 || brew install tmux; \
	elif [ "$(OS)" = "Linux" ]; then \
		echo "Running setup on Linux. Checking for tmux..."; \
		if ! command -v tmux >/dev/null 2>&1; then \
			sudo apt-get update && sudo apt-get install -y tmux; \
		fi; \
	fi
	@echo "Setup complete. You can now run 'make start-db' or 'make serve' etc."

start-db:
	docker run -d --name go_chat_db \
		-e POSTGRES_PASSWORD=${PG_PASS} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_DB=${DB_NAME} \
		-v ${VOLUME_NAME}:/var/lib/postgresql/data \
		-p 5432:5432 \
		postgres:15

stop-db:
	docker stop go_chat_db || true
	docker rm go_chat_db || true
	docker volume rm ${VOLUME_NAME} || true

db: check-os check-docker
	@echo "Recreating database container from scratch..."
	@make stop-db || { echo "Failed to stop/remove DB container."; exit 1; }
	@make start-db || { echo "Failed to start DB container."; exit 1; }
	@echo "Database container recreated successfully."

open-db:
	@tmux kill-session -t open-db 2>/dev/null || true
	@tmux new-session -d -s open_db "\
		docker exec -it go_chat_db \
		psql -U ${DB_USER} -d ${DB_NAME} \
	"

server:
	@tmux kill-session -t server 2>/dev/null || true
	@tmux new-session -d -s server "\
		go run main.go --server \
	"

feed:
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux new-session -d -s feed "\
		go run main.go --client-stream \
	"

publisher:
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux new-session -d -s publisher "\
        go run main.go --publisher; \
    "

client:
	@$(call check-tmux-session,server) 
	@make feed
	@make publisher

status:
	@tmux ls || echo "No tmux sessions running."

a-%:
	tmux attach-session -t $*

check-tmux-session = \
	sleep 1;\
	if ! tmux has-session -t $1 2>/dev/null; then \
		echo "ERROR: Tmux session '$1' was not created."; \
		exit 1; \
	fi

app:
	@make setup || { echo "Setup failed."; exit 1; }
	@make db || { echo "Failed to set up the database."; exit 1; }
	@make server || { echo "Failed to start the server."; exit 1; }
	@$(call check-tmux-session,server) 
	@make client || { echo "Failed to start the client."; exit 1; }
	@$(call check-tmux-session,feed)
	@$(call check-tmux-session,publisher)
	@echo "App successfully started with all components running."

stop:
	@echo "Stopping Docker container..."
	@make stop-db
	@echo "Stopping tmux sessions..."
	@tmux kill-session -t server 2>/dev/null || true
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux kill-session -t feed 2>/dev/null || true
	@tmux kill-session -t open_db 2>/dev/null || true
	@echo "All stopped."

.PHONY: server client