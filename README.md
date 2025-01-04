Near Real-Time Chat with Message Persistence (Go)

This is a simple chat application written in Go. It supports multiple client connections, distributes messages in (near) real-time, and loads persisted chat history at the start of each session.

Features

•	Multiple Client Connections
The server can handle multiple clients at once.

•	Message Broadcasting
When a client sends a message, the server distributes it to all other connected clients (excluding the sender).

•	Message Persistence
Messages are saved so that when a new client connects, they can load the chat history before joining in real-time.

•	PostgreSQL Database
Uses a PostgreSQL container (via Docker) for message persistence.

Requirements
	1.	Go (1.18+ recommended)
	2.	Docker (for running the PostgreSQL database)
	3.	Make (for using the provided Makefile)
	4.	Environment File (.env) with the following variables:
	•	DB_USER
	•	DB_NAME
	•	PG_PASS

Example .env file:

DB_USER=myuser
DB_NAME=go_chat
PG_PASS=secretpass

Getting Started
	1.	Clone the Repository

git clone https://github.com/your-username/your-repo.git


	2.	Navigate to the Project Directory

cd your-repo


	3.	Set Up Your Environment File
	•	Create a .env file in the project root if you haven’t already, and update the variables with the correct credentials.
	4.	Run the Database

make start-db

This command:
	•	Pulls and runs the official postgres:15 Docker image.
	•	Names the container go_chat_db.
	•	Sets up environment variables for the container (user, password, database).
	•	Exposes port 5432 (default PostgreSQL port).
	•	Creates/uses a Docker volume named my_postgres_data.

	5.	Build and Run
	•	Build the Go application:

make back
	•	Run server (backend):

make front
	•	Run a client (front end):



Makefile Commands

Compiles the Go application into an executable named app.
	•	make start-db
Starts a Docker container for PostgreSQL (go_chat_db), with appropriate environment variables and volume mappings.
	•	make stop-db
Stops and removes the running go_chat_db container.
	•	make remove-volume
Removes the Docker volume named my_postgres_data. (Note: This deletes all stored data.)
	•	make db
A convenience command that stops the DB, removes the volume, and starts a fresh PostgreSQL container.
	•	make open-db
Enters the PostgreSQL CLI inside the running container, for manual queries:

psql -U ${DB_USER} -d ${DB_NAME}


	•	make front
Runs go run main.go --client, starting a client interface.
	•	make back
Runs go run main.go --server, starting the chat server.

Usage
	1.	Start the database

make start-db


	2.	Start the server (in one terminal window):

make back


	3.	Open one or more clients (in separate terminals):

make front


	4.	Test the chat by typing messages in each client. You will see the server broadcast messages to all other clients.

Source

This project is inspired by the Realtime Chat Challenge on CodingChallenges.fyi.

Feel free to add sections for Configuration, Architecture, Testing, or anything else that may help users and contributors get the most out of this application.