.PHONY: deploy up backend frontend down restart logs

deploy:
	docker compose up --detach --remove-orphans --build

up:
	docker compose up --detach --remove-orphans --build

backend:
	docker compose up backend --remove-orphans --build

frontend:
	dotnet watch --project .\frontend\BlazorApp\ --no-hot-reload

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f
