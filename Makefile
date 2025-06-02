DB_CONN_URL="pgx5://notfuser:notfpass@localhost:5432/notificationdb"

.PHONY: lint
lint:
	golangci-lint run

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: run-docker
run-docker:
	docker compose down --volumes --remove-orphans 
	docker-compose up --build

.PHONY: test
test:
	ginkgo run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: migrate-up
migrate-up:
	migrate -path ./migrations -database $(DB_CONN_URL) up

.PHONY: migrate-down
migrate-down:
	migrate -path ./migrations -database $(DB_CONN_URL) down