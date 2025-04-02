.PHONY: up backend down restart logs

up:
	docker compose up --detach --remove-orphans --build
	dotnet watch --project .\frontend\BlazorApp\ --no-hot-reload

backend:
	docker compose up --remove-orphans --build

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f
