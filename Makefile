.PHONY: up down restart logs

up:
	docker compose up --detach
	dotnet watch --project .\frontend\BlazorApp\ --no-hot-reload

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f
