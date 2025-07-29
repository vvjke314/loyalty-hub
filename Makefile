volume-rm:
	docker volume rm loyalityhub_loyalityhub_orders_db

container-rm:
	docker rm migrator orders_db

build:
	docker compose build

run:
	docker compose up 