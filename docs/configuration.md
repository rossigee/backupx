# Configuration Guide

backupx supports both YAML and JSON configuration files. The tool automatically detects the format based on file extension.

## File Format

Specify the configuration file as a command-line argument:
```bash
./backupx config.yaml
./backupx backup.json
```

## Configuration Structure

```yaml
sources:
  - id: string                    # Unique identifier
    type: string                  # postgresql, mysql, or folder
    name: string                  # Human-readable name
    [type-specific options...]

destinations:
  - id: string                    # Unique identifier
    type: string                  # Currently only 's3'
    [type-specific options...]

notifications:
  - id: string                    # Unique identifier
    type: string                  # slack, backup-registry, logging, or flagfile
    [type-specific options...]
```

## Sources

### PostgreSQL

```yaml
sources:
  - id: prod_db
    type: postgresql              # or 'pgsql'
    name: Production Database
    host: localhost
    port: 5432
    database: mydb
    user: postgres
    password: secret
    passphrase: encryption-key    # Required for encryption
    compress_only: 0              # 1 = compress without encryption
```

All options except `id`, `type`, and `name` are optional with sensible defaults.

**Default Values:**
- host: localhost
- port: 5432
- user: postgres
- password: (empty)

### MySQL

```yaml
sources:
  - id: mysql_db
    type: mysql                   # or 'mariadb'
    name: MySQL Database
    host: localhost
    port: 3306
    database: mydb
    user: root
    password: secret
    passphrase: encryption-key    # Required for encryption
    compress_only: 0
```

**Default Values:**
- host: localhost
- port: 3306
- user: root
- password: (empty)

### Folder

```yaml
sources:
  - id: code_backup
    type: folder
    name: Source Code
    path: /var/www/myapp
    passphrase: encryption-key    # Required for encryption
    exclude: ".git,.node_modules,*.pyc,__pycache__"  # Comma-separated patterns
```

**Options:**
- `path`: Directory to backup (required)
- `passphrase`: Encryption key (optional)
- `exclude`: Exclude patterns (optional, comma-separated)

## Destinations

### S3

```yaml
destinations:
  - id: s3_primary
    type: s3
    bucket: my-backups            # S3 bucket name (required)
    region: us-east-1             # AWS region (default: us-east-1)
    endpoint: https://s3.amazonaws.com  # Override for non-AWS
    access_key_id: AKIA...        # AWS access key (required)
    secret_access_key: wJal...    # AWS secret key (required)
    retention_copies: 10          # Keep last N backups (optional)
```

**Default Values:**
- region: us-east-1
- endpoint: https://s3.amazonaws.com

**Path Format:**
Backups are stored as: `s3://bucket/hostname/source-id/YYYY-MM-DD/filename.gpg`

Example: `s3://my-backups/web-01/prod_db/2026-09-27/mydb-2026-09-27.sql.gpg`

### Retention Policy

The `retention_copies` setting automatically deletes old backups:
- Keeps only the N most recent backups per source
- Runs after each successful backup
- Ignores files without matching naming scheme

## Notifications

### Slack

```yaml
notifications:
  - id: slack_prod
    type: slack
    url: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"
    notify_on_success: 1          # 1 or 0 (default: 1)
    notify_on_failure: 1          # 1 or 0 (default: 1)
```

**URL Format:**
Generate webhook URL in Slack workspace:
1. Go to https://api.slack.com/apps
2. Create New App → From scratch
3. Enable Incoming Webhooks
4. Add New Webhook to Workspace
5. Copy the Webhook URL

### BackupRegistry

```yaml
notifications:
  - id: registry
    type: backup-registry
    url: "https://backups.example.com/v1/backup-runs"  # Registry endpoint
    token: "bearer-token-here"    # Optional Bearer token
    environment: production       # Custom metadata (optional)
    region: us-east-1            # Custom metadata (optional)
```

**Payload Format:**
```json
{
  "run_id": "550e8400-e29b-41d4-a716-446655440000",
  "job_name": "Production Database",
  "agent_id": "prod_db",
  "start_time": "2026-09-27T10:00:00Z",
  "end_time": "2026-09-27T10:05:00Z",
  "status": "success|failure",
  "bytes_backed_up": 524288000,
  "encrypted": true,
  "encryption_status": "encrypted|compressed|unencrypted",
  "backup_url": "s3://bucket/...backup.gpg",
  "metadata": {"environment": "production", "region": "us-east-1"},
  "error": null
}
```

### Logging

```yaml
notifications:
  - id: log_output
    type: logging               # or 'stdout'
    notify_on_success: 1
    notify_on_failure: 1
```

Outputs backup status to standard logging with structured fields.

### Flag File

```yaml
notifications:
  - id: flag_files
    type: flagfile
    success_path: /var/backups/.backup_success
    failure_path: /var/backups/.backup_failure
    notify_on_success: 1
    notify_on_failure: 1
```

**Behavior:**
- Creates success marker file on successful backup
- Creates failure marker file on failed backup
- File contains timestamp and status
- Useful for monitoring scripts

## Complete Example

```yaml
sources:
  - id: prod_pgsql
    type: postgresql
    name: Production Database
    host: db.example.com
    port: 5432
    database: production
    user: backup_user
    password: ${DB_PASSWORD}
    passphrase: ${BACKUP_PASSPHRASE}

  - id: configs
    type: folder
    name: Application Config
    path: /etc/myapp
    exclude: ".git,*.tmp,node_modules"
    passphrase: ${BACKUP_PASSPHRASE}

destinations:
  - id: aws_s3
    type: s3
    bucket: company-backups
    region: us-east-1
    access_key_id: ${AWS_ACCESS_KEY}
    secret_access_key: ${AWS_SECRET_KEY}
    retention_copies: 30

notifications:
  - id: slack
    type: slack
    url: ${SLACK_WEBHOOK_URL}
    notify_on_success: 1
    notify_on_failure: 1

  - id: registry
    type: backup-registry
    url: https://backups.internal/v1/backup-runs
    token: ${REGISTRY_TOKEN}
    environment: production
    datacenter: us-east-1

  - id: logging
    type: logging
    notify_on_failure: 1
```

## Environment Variables

Configuration can reference environment variables using `${VAR_NAME}` syntax:

```yaml
database:
  password: ${DB_PASSWORD}
  user: ${DB_USER}
```

At runtime:
```bash
export DB_PASSWORD=secret123
export DB_USER=postgres
./backupx config.yaml
```

## Validation

- `sources`: At least one source required
- `destinations`: At least one destination required (or compress_only=1)
- `notifications`: Optional (at least one recommended)
- IDs must be unique within their type
- Required fields must be non-empty

## Format Comparison

### YAML
- Human-readable
- Supports comments
- Recommended for production

### JSON
- Programmatically generated
- Validation friendly
- Use for automation

## Tips

1. **Separate Config for Secrets**
   ```bash
   # config-base.yaml has defaults
   # config-secrets.yaml has sensitive data loaded via env vars
   ```

2. **Use Service Accounts**
   Create dedicated database and S3 users with minimal permissions

3. **Test Configuration**
   ```bash
   ./backupx config.yaml  # Validates and runs backup
   ```

4. **Monitor Logs**
   Add logging notification to catch all issues

5. **Retention Policy**
   Set `retention_copies` to avoid disk bloat
