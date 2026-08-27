.PHONY: docker-up docker-down docker-down-volumes docker-ps docker-logs docker-build

COMPOSE = docker compose -f deployments/docker/docker-compose.yml --env-file deployments/docker/.env

docker-up:
	@test -f deployments/docker/.env || (echo "Missing deployments/docker/.env — run: cp deployments/docker/.env.example deployments/docker/.env" && exit 1)
	$(COMPOSE) up -d --build

docker-build:
	@test -f deployments/docker/.env || (echo "Missing deployments/docker/.env — run: cp deployments/docker/.env.example deployments/docker/.env" && exit 1)
	$(COMPOSE) build

docker-down:
	$(COMPOSE) down

docker-down-volumes:
	$(COMPOSE) down -v

docker-ps:
	$(COMPOSE) ps

docker-logs:
	$(COMPOSE) logs -f --tail=200
