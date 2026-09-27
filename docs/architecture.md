# Architecture Guide

backupx implements a streaming backup pipeline in Go with a focus on efficiency, composability, and reliability.

## Core Design Pattern

The backup process follows a **streaming pipeline** architecture:

```
Source                Encryption              Destinations       Notifications
(Reader)      →        (GPG Pipe)      →      (Uploads)    →     (HTTP/Log/File)
  │                        │                     │                    │
  ├─ PostgreSQL            ├─ AES256             ├─ S3              ├─ Slack
  ├─ MySQL                 └─ Compression       └─ MinIO           ├─ BackupRegistry
  └─ Folder                                                         ├─ Logging
                                                                    └─ Flag Files
```

### Key Principles

1. **Never Buffer Entire Content**: Use `io.Reader`/`io.Writer` interfaces
2. **Fail Fast**: Detect errors early, propagate with context
3. **Streaming**: Memory usage constant regardless of backup size
4. **Composability**: Mix sources, destinations, and notifications
5. **Type Safety**: Factory pattern for runtime instantiation

## Components

### Sources (`sources/*/`)

**Interface:**
```go
type SourceReader interface {
    GetReader() (io.ReadCloser, error)
}
```

**Implementations:**
- `postgresql/`: `pg_dump` with piped compression
- `mysql/`: `mysqldump` with piped compression
- `folder/`: `tar` with gzip compression and exclude patterns

**Design:**
- Launches external command (pg_dump, mysqldump, tar)
- Pipes stdout to Reader
- No temporary files
- Supports exclude patterns (folders only)

### Encryption (`internal/encryption/`)

**GPG Streaming**
```go
type GPGEncryptor struct {
    passphrase   string
    compressOnly bool
}

func (g *GPGEncryptor) Encrypt(input io.Reader) (io.ReadCloser, error)
```

**Features:**
- AES256 symmetric encryption
- Optional compression-only mode
- Streaming: reads input, writes encrypted output
- Uses `--passphrase` flag (safe from pipe issues)

**Example:**
```
Input: 100MB database → GPG → 50MB encrypted output
Memory: ~2MB buffer (not 100MB)
```

### Pipeline (`internal/pipeline/`)

**Orchestration:**
```go
type BackupPipeline struct {
    encryptor    *GPGEncryptor
    destinations []DestinationWriter
    passphrase   string
    sourceId     string
    sourceName   string
    compressOnly bool
    startTime    time.Time
}

func (bp *BackupPipeline) Execute(ctx context.Context, sourceReader io.ReadCloser, filename string) BackupResult
```

**Flow:**
1. Source reader → encrypt → multiple destinations (sequentially)
2. Track success/failure, duration, path
3. Return structured result for notifications

**Error Handling:**
- Encryption failure → return error immediately
- First destination failure → return error, stop pipeline
- No partial uploads on error

### Destinations (`destinations/s3/`)

**Interface:**
```go
type DestinationWriter interface {
    UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error)
}
```

**S3 Implementation:**
- Uses `minio-go` v7.3.0
- Streaming upload (no temp files)
- Automatic bucket creation
- Retention policy enforcement
- Path format: `s3://bucket/hostname/source-id/YYYY-MM-DD/filename.gpg`

**Example:**
```
Input: encrypted reader → S3.UploadReader → s3://bucket/...
Progress: streamed to cloud (constant memory)
```

### Notifications (`notifications/*/`)

**Interface:**
```go
type Notifier interface {
    Notify(ctx context.Context, result BackupResult) error
}
```

**Implementations:**
- `slack/`: HTTP POST with rich formatting
- `backupregistry/`: HTTP POST with Bearer token
- `logging/`: Structured log output
- `flagfile/`: Success/failure marker files

**Configuration:**
```go
notify_on_success: "1"  // Send on success
notify_on_failure: "1"  // Send on failure
```

## Data Flow Example

### Backup of 500MB PostgreSQL Database

```
1. Source (PostgreSQL)
   └─ pg_dump → pipe → 600MB uncompressed

2. Encryption (GPG)
   └─ AES256 encrypt → pipe → 300MB encrypted

3. Destination (S3)
   └─ minio-go → S3 upload → s3://bucket/host/db/2026-09-27/db.sql.gpg

4. Notifications
   ├─ Slack → ✅ Backup successful (500MB in 2.3s)
   ├─ Registry → ✅ Logged run_id + metadata
   └─ Logging → ✅ Structured log output
```

### Memory Usage During Backup

```
Time  Memory
│
├─ 0s: 5MB (initialization)
│
├─ 1s: 25MB (source streaming)
│     └─ input buffer + output buffer
│
├─ 2s: 25MB (encryption streaming)
│     └─ same buffers, data flowing through
│
└─ 3s: 5MB (cleanup, done)

Peak: 25MB (not 500MB of backup data!)
```

## Factory Pattern

Runtime instantiation of sources, destinations, notifications:

```go
func newSource(config Config) interface{} {
    switch config.GetType() {
    case "postgresql":
        return postgresql.NewPostgreSQLSource(config)
    case "mysql":
        return mysql.NewMySQLSource(config)
    case "folder":
        return folder.NewFolderSource(config)
    default:
        return nil, fmt.Errorf("unknown source type")
    }
}
```

**Benefits:**
- Add new sources without modifying main.go
- Configuration drives instantiation
- Type safety with interfaces
- Easy testing with mocks

## Configuration Parsing

```
File Input (YAML/JSON)
    ↓
Config Parser
    ├─ jsonconfig.ParseJSONConfigFile() → IBackupConfig
    └─ yamlconfig.ParseYAMLConfigFile() → IBackupConfig
    ↓
Factory Functions
    ├─ newSource() → SourceReader
    ├─ newDestination() → DestinationWriter
    └─ newNotification() → Notifier
    ↓
Backup Loop
```

## Error Handling Strategy

**Fail-Fast Principle:**

```go
1. Source error (e.g., DB connection)
   └─ Stop immediately, notify with error

2. Encryption error (e.g., invalid passphrase)
   └─ Stop immediately, notify with error

3. Destination error (e.g., S3 upload failed)
   └─ Stop immediately, notify with error

4. Notification error (e.g., Slack webhook down)
   └─ Log, but don't fail backup
```

**Error Context:**
```go
type BackupResult struct {
    Success    bool       // true/false
    Error      string     // Error message if failed
    Message    string     // Additional context
    Duration   int64      // Milliseconds
    BackupPath string     // Path to uploaded file
}
```

## Testing Strategy

### Unit Tests (80+ tests)

```
sources/        → Test config parsing, commands
destinations/   → Test path formatting, uploads
notifications/  → Test config, message formatting
internal/       → Test encryption, pipeline, parsing
```

### Integration Tests

```
Test with MinIO docker container
├─ Full upload workflow
├─ Retention policy
└─ Error handling
```

### E2E Tests

```
Docker Compose setup
├─ Start MinIO
├─ Run full backup
├─ Verify S3 contents
└─ Cleanup
```

See [E2E_TESTING.md](../E2E_TESTING.md) for test coverage details.

## Performance Characteristics

### Time Complexity
- Source: O(n) — read entire source once
- Encryption: O(n) — encrypt streaming
- Upload: O(n) — upload streaming
- **Total: O(n)** — single pass through data

### Space Complexity
- Buffer size: O(1) — constant ~2-4MB buffers
- **Total: O(1)** — independent of backup size

### Network
- **Streaming**: No temporary files on local disk
- **Progress**: Calculated from S3 response
- **Retries**: Delegated to minio-go

## Extensibility

### Adding a New Source Type

1. Create `sources/newtype/newtype.go`
2. Implement `SourceReader` interface
3. Add to `newSource()` factory in main.go
4. Add tests in `sources/newtype/newtype_test.go`

### Adding a New Destination Type

1. Create `destinations/newtype/newtype.go`
2. Implement `DestinationWriter` interface
3. Add to `newDestination()` factory in main.go
4. Add tests in `destinations/newtype/newtype_test.go`

### Adding a New Notification Type

1. Create `notifications/newtype/newtype.go`
2. Implement `Notifier` interface
3. Add to `newNotification()` factory in main.go
4. Add tests in `notifications/newtype/newtype_test.go`

## Dependencies

**Core:**
- Go 1.27.1
- Standard library (no external imports except below)

**External:**
- `github.com/minio/minio-go/v7` — S3 client
- `github.com/google/uuid` — UUID generation for backups
- `gopkg.in/yaml.v2` — YAML parsing

**Development:**
- `github.com/golangci/golangci-lint/v2` — Code quality
- Docker Compose — E2E testing

## Deployment Considerations

### Security
1. **Encryption**: AES256 via GPG, passphrases via config
2. **Credentials**: Use environment variables, not files
3. **Audit**: Structured logging with all operations

### Reliability
1. **Streaming**: No data loss from crashes
2. **Retention**: Automatic cleanup of old backups
3. **Notifications**: Alert on all failures

### Performance
1. **Concurrency**: Single-threaded (prevent resource contention)
2. **Timeouts**: 10s for HTTP requests, configurable for DBs
3. **Retries**: Built-in via minio-go (AWS SDK)

## Future Enhancements

- [ ] Parallel uploads to multiple destinations
- [ ] Incremental backups (diff-based)
- [ ] Database-native encrypted transfer
- [ ] Metrics/Prometheus export
- [ ] Web UI for monitoring
- [ ] Backup verification/restore testing
