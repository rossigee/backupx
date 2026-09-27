# AGENTS.md — Development Guide for AI Agents

This guide helps AI agents (Claude, others) understand backupx codebase structure, patterns, and development workflows.

## Project Overview

**backupx** is a Go backup tool with streaming pipeline architecture.

```
CLI Entry    Config Parse    Factory         Pipeline        S3 Upload
  ↓             ↓               ↓              ↓                ↓
main.go  → jsonconfig    → newSource()  → pipeline    → destinations/s3
          yamlconfig        newDest()      execute()
                            newNotif()
```

**Key Principle**: Stream data through pipeline without buffering.

## Directory Structure

```
backupx/
├── main.go                          # CLI entry, orchestration
├── go.mod / go.sum                  # Dependencies
├── Makefile                         # Build targets
├── docker-compose.test.yml          # MinIO for testing
│
├── sources/                         # Backup source implementations
│   ├── postgresql/postgresql.go      # pg_dump wrapper
│   ├── mysql/mysql.go               # mysqldump wrapper
│   └── folder/folder.go             # tar wrapper
│
├── destinations/                    # Upload destination implementations
│   └── s3/s3.go                     # minio-go S3 client
│
├── notifications/                   # Notification implementations
│   ├── backupregistry/              # HTTP POST (generic backup API)
│   ├── slack/                       # Slack webhook
│   ├── logging/                     # Structured logging
│   └── flagfile/                    # File marker
│
├── internal/                        # Shared internal packages
│   ├── encryption/gpg.go            # GPG streaming encryption
│   ├── pipeline/backup.go           # Main backup orchestration
│   ├── jsonconfig/                  # JSON parser
│   └── yamlconfig/                  # YAML parser
│
├── docs/                            # User documentation
│   ├── CONFIGURATION.md             # Config file guide
│   └── ARCHITECTURE.md              # Design deep-dive
│
├── .github/workflows/               # CI/CD
│   ├── test.yml                     # Tests on push/PR
│   └── release.yml                  # Builds on tag
│
└── README.md / AGENTS.md            # User & developer docs
```

## Key Files & Responsibilities

### main.go (250+ lines)
**Purpose**: CLI entry point, configuration loading, source/destination/notification orchestration

**Key Functions**:
- `main()`: Entry point, backup loop
- `parseArgs()`: CLI flag parsing
- `newSource()`: Factory for sources (PostgreSQL, MySQL, Folder)
- `newDestination()`: Factory for destinations (S3)
- `newNotification()`: Factory for notifications (Slack, BackupRegistry, etc.)

**Pattern**: Factory functions take `interface{}` config, return typed instances

**Editing**: When adding new source/destination/notification type:
1. Add case to appropriate `new*()` function
2. Call type-specific init function
3. Example: `case "postgresql": return initPostgreSQLSource(config)`

### internal/pipeline/backup.go (115 lines)
**Purpose**: Main backup orchestration — chains source → encryption → destinations

**Key Struct**:
```go
type BackupPipeline struct {
    encryptor    *encryption.GPGEncryptor
    destinations []DestinationWriter
    sourceId, sourceName string
    startTime    time.Time
}
```

**Execute() Flow**:
1. Take `io.ReadCloser` from source
2. Encrypt via GPG
3. Upload to each destination sequentially
4. Return `BackupResult` struct (success, error, path, duration)

**Key Pattern**: All operations streaming (no buffering), defer cleanup on errors

### internal/encryption/gpg.go (66 lines)
**Purpose**: GPG AES256 encryption without temporary files

**Key Method**:
```go
func (g *GPGEncryptor) Encrypt(input io.Reader) (io.ReadCloser, error)
```

**Pattern**: Wraps `gpg` command, pipes stdin/stdout, returns reader

**Important**: Uses `--passphrase` flag (not stdin), handles both encryption and compression-only modes

### sources/*/source.go (50-70 lines each)
**Pattern**: Implement `SourceReader` interface
```go
type SourceReader interface {
    GetReader() (io.ReadCloser, error)
}
```

**Examples**:
- `postgresql/`: Exec `pg_dump`, return stdout
- `mysql/`: Exec `mysqldump`, return stdout
- `folder/`: Exec `tar`, return stdout

**Common Pattern**: External command → pipe stdout → return as reader

### destinations/s3/s3.go (180 lines)
**Purpose**: S3 upload via minio-go

**Key Method**:
```go
func (s *S3Destination) UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error)
```

**Pattern**: Stream reader directly to S3 (no buffering), return uploaded path

**Features**:
- Automatic bucket creation
- Retention policy (delete old backups)
- Path formatting: `s3://bucket/hostname/source-id/YYYY-MM-DD/filename.gpg`

### notifications/*/notification.go (50-130 lines each)
**Pattern**: Implement `Notifier` interface
```go
type Notifier interface {
    Notify(ctx context.Context, result BackupResult) error
}
```

**Input**: `BackupResult` struct with success, error, path, duration

**Examples**:
- `slack/`: HTTP POST with rich formatting
- `backupregistry/`: HTTP POST with Bearer token + UUID
- `logging/`: Structured log output
- `flagfile/`: Write marker files

## Code Patterns

### Error Handling
**Pattern**: Return early, fail fast
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Resource Cleanup
**Pattern**: Defer with explicit error ignore
```go
defer func() {
    _ = resource.Close()
}()
```

Not just `defer resource.Close()` (linter error).

### Interface-Based Design
**Pattern**: Use small interfaces, factory functions
```go
type Reader interface {
    GetReader() (io.ReadCloser, error)
}

func newSource(config Config) interface{} {
    switch config.Type {
    case "postgresql":
        return postgresql.NewPostgreSQLSource(config)
    // ...
    }
}
```

**Benefit**: Add new types without modifying main.go

### Streaming (Critical)
**Pattern**: Never call `io.ReadAll()` on large data
```go
// WRONG: buffers entire backup
data, _ := io.ReadAll(reader)
// upload(data)

// RIGHT: streams
encrypt.Encrypt(reader)  // returns another reader
s3.Upload(encryptedReader)  // streams to S3
```

## Testing

### Test Files
```
*_test.go files are in same package as code being tested
Example: sources/postgresql/postgresql_test.go
```

### Test Patterns

**Unit Test**: Test config parsing, no external dependencies
```go
func TestNewPostgreSQLSourceMissingDatabase(t *testing.T) {
    config := testConfig{...}
    _, err := postgresql.NewPostgreSQLSource(config)
    if err == nil {
        t.Error("Expected error")
    }
}
```

**Integration Test**: Test with real S3 (MinIO in CI)
```go
func TestS3Upload(t *testing.T) {
    // Creates real S3 connection
    // Uploads real data
    // Verifies on S3
}
```

**Table-Driven Test**: Multiple scenarios
```go
tests := []struct{
    name string
    input string
    expected string
}{
    {"case1", "input1", "expected1"},
    {"case2", "input2", "expected2"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {...})
}
```

### Running Tests
```bash
make test           # All tests
make test-unit      # Fast unit only
make test-e2e       # With MinIO (requires Docker)
make lint           # golangci-lint
```

## Development Workflow

### Adding a New Source Type

**Step 1**: Create `sources/newtype/newtype.go`
```go
package newtype

type NewTypeSource struct {
    config Config
}

func NewNewTypeSource(cfg Config) (*NewTypeSource, error) {
    // validate config
    return &NewTypeSource{config: cfg}, nil
}

func (n *NewTypeSource) GetReader() (io.ReadCloser, error) {
    // Execute external command or connect to service
    // Return reader to backup stream
}
```

**Step 2**: Add tests `sources/newtype/newtype_test.go`
```go
func TestNewNewTypeSourceMissingRequired(t *testing.T) { ... }
func TestNewNewTypeSourceDefaults(t *testing.T) { ... }
```

**Step 3**: Register in main.go
```go
func newSource(config Config) interface{} {
    switch config.Type {
    // ...
    case "newtype":
        return initNewTypeSource(config)
    }
}

func initNewTypeSource(cfg Config) (interface{}, error) {
    return newtype.NewNewTypeSource(cfg)
}
```

**Step 4**: Update `docs/configuration.md` with new source docs

### Adding a New Notification Type

**Similar to sources**:
1. Create `notifications/newtype/newtype.go` with `Notify(ctx, result) error`
2. Add tests
3. Register in main.go `newNotification()`
4. Document in `docs/configuration.md`

### Editing Existing Code

**Before editing**:
1. Read the file completely
2. Understand the pattern (factory, streaming, error handling)
3. Check related tests to understand expected behavior

**When editing**:
1. Maintain existing patterns
2. Add/update tests for changes
3. Run `make lint` before submitting
4. Run `make test` to verify

**Common edits**:
- Adding config options: Update struct, parse function, tests, docs
- Fixing bugs: Add test that fails with bug, fix code, test passes
- Performance: Profile first, measure improvements

## Common Tasks

### Task: Add new config option to PostgreSQL source

1. Edit `sources/postgresql/postgresql.go`:
   ```go
   type PostgreSQLSource struct {
       // ... existing fields
       NewOption string  // ADD
   }
   ```

2. Update `NewPostgreSQLSource()` to parse it:
   ```go
   newOption := attrs["new_option"]
   ```

3. Use it in `GetReader()` if needed

4. Add test in `sources/postgresql/postgresql_test.go`:
   ```go
   func TestNewOptionParsing(t *testing.T) {
       config := testConfig{attrs: map[string]string{"new_option": "value"}}
       src, _ := NewPostgreSQLSource(config)
       if src.NewOption != "value" {
           t.Error("Option not parsed")
       }
   }
   ```

5. Update `docs/configuration.md` with new option

6. Run `make test` and `make lint`

### Task: Debug a failing test

1. Run specific test: `go test ./sources/postgresql -run TestName -v`
2. Add debug output: `t.Logf("value: %v", variable)`
3. Check test expectations vs actual behavior
4. Add assertions to catch edge cases

### Task: Optimize memory usage

1. Profile with pprof: `go test -cpuprofile=cpu.prof ./...`
2. Identify buffering: Look for `io.ReadAll()`, `ioutil.WriteFile()`, etc.
3. Replace with streaming: Use readers/writers directly
4. Test improvement: Compare memory before/after

## Dependencies

**External**:
- `github.com/minio/minio-go/v7` — S3 client
- `github.com/google/uuid` — UUID generation
- `gopkg.in/yaml.v2` — YAML parsing

**Development**:
- Go 1.27.1
- golangci-lint v2.14.0
- Docker Compose (for MinIO in tests)

**Constraints**:
- No Windows-specific code (use Go stdlib)
- No cgo dependencies (keeps builds simple)
- No package dependencies not in go.mod

## Common Mistakes to Avoid

❌ **Buffering entire backup**: Don't use `io.ReadAll()` on large readers
✅ **Instead**: Pass reader through pipeline

❌ **Ignoring errors**: Don't use `defer resource.Close()` without wrapping
✅ **Instead**: Use `defer func() { _ = resource.Close() }()`

❌ **Adding features without tests**: Every change needs tests
✅ **Instead**: Write test first, then implementation

❌ **Modifying main.go without factory pattern**: Don't hardcode types
✅ **Instead**: Use factory functions, allow new types via config

❌ **Assuming localhost**: Don't hardcode endpoints
✅ **Instead**: Use configuration, test with environment variables

❌ **Skipping documentation**: Users and agents need to understand config
✅ **Instead**: Update docs whenever adding options

## Performance Guidelines

- **Peak Memory**: <50MB (streaming, not buffering)
- **Time Complexity**: O(n) single pass
- **Space Complexity**: O(1) constant buffers
- **CPU**: Single-threaded, parallel in pipeline jobs

**Measure**: Always profile before optimizing

## Release Process

1. **Ensure tests pass**: `make test`
2. **Check linting**: `make lint`
3. **Update version**: Document in code comments
4. **Create tag**: `git tag v1.0.0`
5. **Push tag**: `git push origin v1.0.0`
6. **GitHub Actions**: Automatically builds and releases

## Getting Help

- **Architecture**: See `docs/architecture.md`
- **Configuration**: See `docs/configuration.md`
- **Testing**: See `docs/testing.md`
- **Code**: Read test files for usage examples

## Key Principles (Remember)

1. **Stream**: Don't buffer entire backups
2. **Fail Fast**: Return errors immediately with context
3. **Test**: Every change needs tests
4. **Document**: Users need to understand configuration
5. **Clean**: Keep code simple, follow Go conventions
