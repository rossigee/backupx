GOLANGCI_LINT_VERSION := v2.14.0

backups:
	go build

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

test:
	go test ./... -v

test-unit:
	go test ./... -v -short

test-e2e:
	@echo "Starting MinIO for e2e testing..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for MinIO to be ready..."
	@for i in 1 2 3 4 5 6; do \
		if docker compose -f docker-compose.test.yml exec -T minio curl -f http://localhost:9000/minio/health/live >/dev/null 2>&1; then \
			echo "MinIO is ready"; \
			break; \
		fi; \
		if [ $$i -lt 6 ]; then echo "Waiting... ($$i/5)"; sleep 2; fi; \
	done
	@echo "Running e2e tests..."
	MINIO_ENDPOINT=http://localhost:9000 MINIO_USER=minioadmin MINIO_PASS=minioadmin go test ./destinations/s3 -v -run TestUpload
	@echo "Stopping MinIO..."
	docker compose -f docker-compose.test.yml down

test-e2e-interactive:
	@echo "Starting MinIO for interactive e2e testing..."
	docker compose -f docker-compose.test.yml up -d
	@echo "MinIO is running at http://localhost:9001 (console)"
	@echo "MinIO S3 endpoint: http://localhost:9000"
	@echo "Credentials: minioadmin / minioadmin"
	@echo "Running e2e tests..."
	MINIO_ENDPOINT=http://localhost:9000 MINIO_USER=minioadmin MINIO_PASS=minioadmin go test ./destinations/s3 -v
	@echo "MinIO is still running. Press Ctrl+C to stop or run: docker compose -f docker-compose.test.yml down"

.PHONY: backups lint test test-unit test-e2e test-e2e-interactive
