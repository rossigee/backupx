# End-to-End Testing Guide

This guide explains how to run the e2e test suite for the backupx S3 destination.

## Quick Start

### Run all tests (unit + integration skipped):
```bash
make test
```

### Run only unit tests (fast):
```bash
make test-unit
```

### Run full e2e test suite with MinIO:
```bash
make test-e2e
```

This will:
1. Start a MinIO instance in Docker
2. Wait for it to be ready
3. Run all S3 destination tests against real MinIO
4. Stop MinIO and clean up

## Interactive E2E Testing

For development and debugging:
```bash
make test-e2e-interactive
```

This starts MinIO and leaves it running so you can:
- Access MinIO console at http://localhost:9001
- Inspect uploaded files manually
- Run tests multiple times without container restart
- Credentials: `minioadmin` / `minioadmin`

Stop it with:
```bash
docker-compose -f docker-compose.test.yml down
```

## Manual E2E Testing

### 1. Start MinIO manually:
```bash
docker-compose -f docker-compose.test.yml up -d
```

### 2. Wait for it to be ready:
```bash
docker-compose -f docker-compose.test.yml exec minio curl http://localhost:9000/minio/health/live
```

### 3. Run tests with MinIO environment:
```bash
MINIO_ENDPOINT=http://localhost:9000 \
MINIO_USER=minioadmin \
MINIO_PASS=minioadmin \
go test ./destinations/s3 -v
```

### 4. Clean up:
```bash
docker-compose -f docker-compose.test.yml down
```

## Test Environment Variables

When running tests, you can override the MinIO endpoint and credentials:

- `MINIO_ENDPOINT` - MinIO S3 endpoint (default: `http://localhost:9000`)
- `MINIO_USER` - Access key ID (default: `minioadmin`)
- `MINIO_PASS` - Secret access key (default: `minioadmin`)

Example with AWS S3:
```bash
MINIO_ENDPOINT=s3.amazonaws.com \
MINIO_USER=your-aws-key \
MINIO_PASS=your-aws-secret \
go test ./destinations/s3 -v -run TestUploadDataIntegrity
```

## What Gets Tested

### Unit Tests (always run):
- ✅ S3 destination initialization
- ✅ Configuration validation
- ✅ Default values
- ✅ Retention policy configuration
- ✅ Error handling

### Integration Tests (run when S3 is available):
- ✅ Path formatting (prefix/hostname/date/filename)
- ✅ Upload data integrity
- ✅ Handling of custom endpoints
- ✅ Bucket creation and cleanup
- ✅ Retention policy enforcement

### Skipped Tests (when S3 unavailable):
- Tests gracefully skip if S3 endpoint is unreachable
- Allows CI/CD without external dependencies
- Full tests run in e2e mode with MinIO

## Troubleshooting

### MinIO fails to start:
```bash
docker-compose -f docker-compose.test.yml logs minio
```

### Tests timeout waiting for MinIO:
Increase the wait time by editing `Makefile` (currently 30 seconds total)

### Port 9000/9001 already in use:
```bash
docker-compose -f docker-compose.test.yml down
# or
lsof -ti:9000,9001 | xargs kill -9
```

### Tests still skip S3:
Verify MinIO is healthy:
```bash
curl http://localhost:9000/minio/health/live
```

Should return HTTP 200 with "OK"

## CI/CD Integration

For GitHub Actions or similar:
```yaml
- name: Run e2e tests
  run: make test-e2e
```

Or for unit tests only:
```yaml
- name: Run unit tests
  run: make test-unit
```
