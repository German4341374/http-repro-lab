SHELL := /bin/sh
.DEFAULT_GOAL := help

.PHONY: help setup format lint typecheck test build integration verify up down clean
help:
	@printf '%s\n' 'setup format lint typecheck test build integration verify up down clean'

setup:
	cd apps/api && npm ci
	python3 -m venv tools/request-minimizer-python/.venv
	tools/request-minimizer-python/.venv/bin/pip install -e 'tools/request-minimizer-python[dev]'

format:
	gofmt -w cmd internal
	cd apps/api && npm run format

lint:
	test -z "$$(gofmt -l cmd internal)"
	go vet ./cmd/... ./internal/...
	cd apps/api && npm run format:check && npm run lint
	tools/request-minimizer-python/.venv/bin/ruff check tools/request-minimizer-python

typecheck:
	cd apps/api && npm run typecheck
	cd tools/request-minimizer-python && .venv/bin/mypy

test:
	go test -race ./cmd/... ./internal/...
	cd apps/api && npm test
	tools/request-minimizer-python/.venv/bin/pytest tools/request-minimizer-python
	dotnet test runners/dotnet-repro-runner/tests/HttpRepro.Runner.Tests --configuration Release
	mvn -B -f runners/java-repro-runner/pom.xml verify
	cd runners/php-repro-runner && composer install --no-interaction && composer test && composer analyse

build:
	go build ./cmd/http-repro ./cmd/mock-api
	cd apps/api && npm run build
	docker build --target cli -t http-repro-lab:local .

integration:
	./scripts/integration.sh

verify: lint typecheck test integration build

up:
	docker compose up --build -d

down:
	docker compose down

clean:
	docker compose down -v --remove-orphans
	rm -rf .repro-workspace dist report generated
