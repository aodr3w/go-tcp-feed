SHELL := /bin/bash

# Include .env if present
-include .env

# OS check (Darwin is macOS)
OS := $(shell uname -s)
APPROVED_OS := Darwin

# Utility targets
check-os:
	@if [ "$(OS)" != "$(APPROVED_OS)" ]; then \
	  echo "ERROR: This Makefile is intended for macOS (Darwin). Found '$(OS)'."; \
	  exit 1; \
	fi

check-docker:
	@which docker > /dev/null 2>&1 || ( echo "ERROR: Docker is not installed. Please install Docker." && exit 1 )
	@docker info > /dev/null 2>&1 || ( echo "ERROR: Docker daemon is not running. Please start Docker." && exit 1 )

# Setup command:
#    - Checks OS
#    - Checks Docker + daemon
#    - Installs tmux if missing
setup: check-os check-docker
	@echo "Running setup on macOS. Checking for tmux..."
	@brew list tmux >/dev/null 2>&1 || brew install tmux
	@echo "Setup complete. You can now run 'make start-db' or 'make serve' etc."

# Start database WITHOUT tmux (for a simpler debug approach).
#    - If you truly want it in a tmux session, uncomment the tmux lines.
start-db:
	docker run -d --name go_chat_db \
		-e POSTGRES_PASSWORD=${PG_PASS} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_DB=${DB_NAME} \
		-v ${VOLUME_NAME}:/var/lib/postgresql/data \
		-p 5432:5432 \
		postgres:15

# Stop and remove database container & volume (ignore errors if not present).
stop-db:
	docker stop go_chat_db || true
	docker rm go_chat_db || true
	docker volume rm ${VOLUME_NAME} || true

# Recreate database container from scratch using start-db and stop-db
db: check-os check-docker
	@echo "Recreating database container from scratch..."
	@make stop-db || { echo "Failed to stop and remove the database container."; exit 1; }
	@make start-db || { echo "Failed to start the database container."; exit 1; }
	@echo "Database container (re)created successfully."

# Open a psql session in a new tmux session
open-db:
	@tmux kill-session -t open-db 2>/dev/null || true
	@tmux new-session -d -s open_db "\
		docker exec -it go_chat_db \
		psql -U ${DB_USER} -d ${DB_NAME} \
	"

# Serve command in its own tmux session
server:
	@tmux kill-session -t server 2>/dev/null || true
	@tmux new-session -d -s server "\
		go run main.go --server \
	"


# Run client stream in separate tmux session
feed:
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux new-session -d -s feed "\
		go run main.go --client-stream \
	"

# Run publisher in a separate tmux session
publisher:
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux new-session -d -s publisher "\
        go run main.go --publisher; \
    "

# client - starts both publisher and client stream in separate tmux sessions
client:
	@$(call check-tmux-session,server) 
	@make feed
	@make publisher

# Combined 'status' to list tmux sessions
status:
	@tmux ls || echo "No tmux sessions running."

# Attach to a named tmux session
a-%:
	tmux attach-session -t $*

# Function to check if a tmux session exists
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
	@$(call check-tmux-session,serve) # Ensure the server session exists
	@make client || { echo "Failed to start the client."; exit 1; }
	@$(call check-tmux-session,feed) # Ensure the client session exists
	@$(call check-tmux-session,publisher) # Ensure the publisher session exists
	@echo "App successfully started with all components running."

# Stop everything: db container + all tmux sessions
stop:
	@echo "Stopping Docker container..."
	@make stop-db
	@echo "Stopping tmux sessions (if running): server, publisher, feed, open_db..."
	@tmux kill-session -t server 2>/dev/null || true
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux kill-session -t feed 2>/dev/null || true
	@tmux kill-session -t open_db 2>/dev/null || true
	@echo "All stopped."

.PHONY: server client