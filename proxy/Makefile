ENV_EXAMPLE_FILES := $(wildcard docker/.env.*.example)
DOCKER_COMPOSE := docker compose -p vox-proxy --env-file docker/.env.development -f docker/docker-compose.yml

.PHONY: up down env

up: env
	$(DOCKER_COMPOSE) up --build --force-recreate -d

down:
	$(DOCKER_COMPOSE) down

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
