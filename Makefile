TEMPLATE_DIR=template
include .env

VOLUME_NAME=my_postgres_data

build:
	go build -o app main.go

start-db:
	docker run -d --name go_chat_db \
	-e POSTGRES_PASSWORD=${PG_PASS} \
	-e POSTGRES_USER=${DB_USER} \
	-e POSTGRES_DB=${DB_NAME} \
	-v ${VOLUME_NAME}:/var/lib/postgresql/data \
	-p 5432:5432 \
	postgres:15

stop-db:
	docker stop go_chat_db && docker rm go_chat_db
	@MAKE remove-volume

remove-volume:
	docker volume rm ${VOLUME_NAME}

open-db:
	docker exec -it go_chat_db \
	psql -U ${DB_USER} -d ${DB_NAME}


db:
	@MAKE stop-db
	@MAKE remove-volume
	@MAKE start-db

front:
	go run main.go --client

back:
	go run main.go --server