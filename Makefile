.PHONY: prod prod-down dev dev-down test test-down

prod:
	docker compose -f docker-compose.yaml up -d --build

prod-down:
	docker compose -f docker-compose.yaml down

dev:
	docker compose -f docker-compose.dev.yaml up

dev-down:
	docker compose -f docker-compose.dev.yaml down

test:
	./backend/scripts/test.sh

test-down:
	docker compose -f docker-compose.test.yaml down -v
