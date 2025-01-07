SHELL := /bin/bash

# Include .env if present
-include .env

# 1. Default variables (in case they're not set in the .env file)
VOLUME_NAME ?= my_postgres_data
PG_PASS ?= postgres
DB_USER ?= postgres
DB_NAME ?= go_chat_db

# 2. OS check (Darwin is macOS)
OS := $(shell uname -s)
APPROVED_OS := Darwin

# 3. Utility targets
check-os:
	@if [ "$(OS)" != "$(APPROVED_OS)" ]; then \
	  echo "ERROR: This Makefile is intended for macOS (Darwin). Found '$(OS)'."; \
	  exit 1; \
	fi

check-docker:
	@which docker > /dev/null 2>&1 || ( echo "ERROR: Docker is not installed. Please install Docker." && exit 1 )
	@docker info > /dev/null 2>&1 || ( echo "ERROR: Docker daemon is not running. Please start Docker." && exit 1 )

# 4. Setup command:
#    - Checks OS
#    - Checks Docker + daemon
#    - Installs tmux if missing
setup: check-os check-docker
	@echo "Running setup on macOS. Checking for tmux..."
	@brew list tmux >/dev/null 2>&1 || brew install tmux
	@echo "Setup complete. You can now run 'make start-db' or 'make serve' etc."

# 5. Start database WITHOUT tmux (for a simpler debug approach).
#    - If you truly want it in a tmux session, uncomment the tmux lines.
start-db:
	docker run -d --name go_chat_db \
		-e POSTGRES_PASSWORD=${PG_PASS} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_DB=${DB_NAME} \
		-v ${VOLUME_NAME}:/var/lib/postgresql/data \
		-p 5432:5432 \
		postgres:15

# 6. Stop and remove database container & volume (ignore errors if not present).
stop-db:
	docker stop go_chat_db || true
	docker rm go_chat_db || true
	docker volume rm ${VOLUME_NAME} || true

# 7. Recreate database container from scratch.
db: check-os check-docker
	echo "VOLUME_NAME=${VOLUME_NAME}"
	echo "PG_PASS=${PG_PASS}"
	echo "DB_USER=${DB_USER}"
	echo "DB_NAME=${DB_NAME}"
	docker stop go_chat_db || true
	docker rm go_chat_db || true
	docker volume rm ${VOLUME_NAME} || true
	docker run -d --name go_chat_db \
		-e POSTGRES_PASSWORD=${PG_PASS} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_DB=${DB_NAME} \
		-v ${VOLUME_NAME}:/var/lib/postgresql/data \
		-p 5432:5432 \
		postgres:15
	@echo "Database container (re)created."

# 8. Open a psql session in a new tmux session
open-db:
	tmux new-session -d -s open_db "\
		docker exec -it go_chat_db \
		psql -U ${DB_USER} -d ${DB_NAME} \
	"

# 9. Serve command in its own tmux session
serve:
	tmux new-session -d -s serve "\
		make db && \
		go run main.go --server \
	"

# 10. Start the backend (alias for 'serve')
backend:
	tmux new-session -d -s backend "make serve"

# 11. Run client stream in separate tmux session
cstream:
	tmux new-session -d -s cstream "\
		go run main.go --client-stream \
	"

# 12. Run publisher in a separate tmux session
publisher:
	tmux new-session -d -s publisher "\
		go run main.go --publisher \
	"

# 13. Frontend - starts both publisher and client stream.
# 13. Frontend - starts both publisher and client stream in separate tmux sessions
frontend:
	@make publisher
	@make cstream

# 14. Combined 'status' to list tmux sessions
status:
	@tmux ls || echo "No tmux sessions running."

# 15. Attach to a named tmux session
attach-%:
	tmux attach-session -t $*

# 16. Stop everything: db container + all tmux sessions
stop:
	@echo "Stopping Docker container..."
	@make stop-db
	@echo "Stopping tmux sessions (if running): serve, backend, publisher, cstream, frontend, open_db..."
	@tmux kill-session -t serve 2>/dev/null || true
	@tmux kill-session -t backend 2>/dev/null || true
	@tmux kill-session -t publisher 2>/dev/null || true
	@tmux kill-session -t cstream 2>/dev/null || true
	@tmux kill-session -t frontend 2>/dev/null || true
	@tmux kill-session -t open_db 2>/dev/null || true
	@echo "All stopped."