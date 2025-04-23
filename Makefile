.PHONY: up backend frontend down restart logs

up:
	docker compose up --detach --remove-orphans --build

backend:
	docker compose up --remove-orphans --build

frontend:
	dotnet watch --project .\frontend\BlazorApp\ --no-hot-reload

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f
