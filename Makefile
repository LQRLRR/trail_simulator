start:
	docker-compose up --build -d
	docker exec -it trail_simulator bash
up:
	docker-compose up -d
build:
	docker-compose build
down:
	docker-comopse down
exec:
	docker exec -it trail_simulator bash