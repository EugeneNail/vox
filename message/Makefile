ENV_EXAMPLE_FILES := $(wildcard docker/.env.*.example)
DOCKER_COMPOSE := docker compose --env-file docker/.env.development -f docker/docker-compose.yml
MIGRATOR_IMAGE := vox-message-migrator:latest
MIGRATOR_DOCKER_RUN := docker run --rm --env-file docker/.env.development --network vox-message-local-network -v $(CURDIR)/migrations:/migrations

.PHONY: up down env migrate rollback

up: env
	$(DOCKER_COMPOSE) up --build -d

down:
	$(DOCKER_COMPOSE) down

migrate: env
	docker build -f docker/migrator.Dockerfile -t $(MIGRATOR_IMAGE) .
	$(MIGRATOR_DOCKER_RUN) $(MIGRATOR_IMAGE)

rollback: env
	docker build -f docker/migrator.Dockerfile -t $(MIGRATOR_IMAGE) .
	$(MIGRATOR_DOCKER_RUN) -e MIGRATOR_COMMAND=down $(MIGRATOR_IMAGE)

env:
	@for example in $(ENV_EXAMPLE_FILES); do \
		target=$${example%.example}; \
		if [ ! -f "$$target" ]; then \
			cp "$$example" "$$target"; \
			continue; \
		fi; \
		while IFS= read -r line || [ -n "$$line" ]; do \
			case "$$line" in \
				''|\#*) continue ;; \
			esac; \
			key=$${line%%=*}; \
			if ! grep -qE "^$${key}=" "$$target"; then \
				printf '\n%s\n' "$$line" >> "$$target"; \
			fi; \
		done < "$$example"; \
	done
