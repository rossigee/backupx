package pipeline

import (
	"bytes"
	"context"
	"io"
	"testing"
)

type recordingDestination struct {
	filename   string
	dataSize   int64
	uploadedOk bool
}

func (r *recordingDestination) UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	r.filename = filename
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	r.dataSize = int64(len(data))
	r.uploadedOk = true
	return "s3://bucket/" + filename, nil
}

func TestBackupPipelineEncryptionPassphrase(t *testing.T) {
	config := PipelineConfig{
		SourceId:     "test-db",
		SourceName:   "Test Database",
		Passphrase:   "secretkey",
		CompressOnly: false,
		Destinations: []DestinationWriter{},
	}

	pipeline := NewBackupPipeline(config)
	data := []byte("sensitive database backup data")
	reader := io.NopCloser(bytes.NewReader(data))

	result := pipeline.Execute(context.Background(), reader, "backup.sql")

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if !bytes.Contains([]byte(result.BackupPath), []byte(".gpg")) {
		t.Errorf("Expected .gpg extension, got %s", result.BackupPath)
	}
}

func TestBackupPipelineWithMultipleDestinations(t *testing.T) {
	dest1 := &recordingDestination{}
	dest2 := &recordingDestination{}

	config := PipelineConfig{
		SourceId:     "test-src",
		SourceName:   "Test Source",
		Passphrase:   "secret",
		CompressOnly: true,
		Destinations: []DestinationWriter{dest1, dest2},
	}

	pipeline := NewBackupPipeline(config)
	data := []byte("backup data")
	reader := io.NopCloser(bytes.NewReader(data))

	result := pipeline.Execute(context.Background(), reader, "backup.tar")

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	if dest1.filename != "backup.tar.gz" {
		t.Errorf("Expected dest1 filename 'backup.tar.gz', got '%s'", dest1.filename)
	}

	if dest2.filename == "" {
		t.Error("Expected dest2 to be called, but filename is empty")
	}

	if result.BackupPath != "s3://bucket/backup.tar.gz" {
		t.Errorf("Expected path 's3://bucket/backup.tar.gz', got '%s'", result.BackupPath)
	}
}

func TestBackupPipelineCompressionOnlyExtension(t *testing.T) {
	dest := &recordingDestination{}

	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test",
		Passphrase:   "",
		CompressOnly: true,
		Destinations: []DestinationWriter{dest},
	}

	pipeline := NewBackupPipeline(config)
	reader := io.NopCloser(bytes.NewReader([]byte("test")))

	_ = pipeline.Execute(context.Background(), reader, "data.tar")

	if dest.filename != "data.tar.gz" {
		t.Errorf("Expected 'data.tar.gz', got '%s'", dest.filename)
	}
}

func TestBackupPipelineEncryptionExtension(t *testing.T) {
	dest := &recordingDestination{}

	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test",
		Passphrase:   "secret",
		CompressOnly: false,
		Destinations: []DestinationWriter{dest},
	}

	pipeline := NewBackupPipeline(config)
	reader := io.NopCloser(bytes.NewReader([]byte("test")))

	result := pipeline.Execute(context.Background(), reader, "data.tar")

	if dest.filename != "data.tar.gpg" {
		t.Errorf("Expected 'data.tar.gpg', got '%s'", dest.filename)
	}

	if !result.Success {
		t.Errorf("Expected success, got: %v", result.Error)
	}
}

func TestBackupPipelineDestinationFailure(t *testing.T) {
	failingDest := &failingDestination{}

	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test",
		Passphrase:   "secret",
		CompressOnly: false,
		Destinations: []DestinationWriter{failingDest},
	}

	pipeline := NewBackupPipeline(config)
	reader := io.NopCloser(bytes.NewReader([]byte("test data")))

	result := pipeline.Execute(context.Background(), reader, "backup.sql")

	if result.Success {
		t.Error("Expected failure when destination fails")
	}

	if result.Error == "" {
		t.Error("Expected error message, got empty string")
	}
}

type failingDestination struct{}

func (f *failingDestination) UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	_, _ = io.ReadAll(reader)
	return "", errUploadFailed
}

var errUploadFailed = io.EOF

func TestBackupPipelineDurationTracking(t *testing.T) {
	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test",
		Passphrase:   "",
		CompressOnly: true,
		Destinations: []DestinationWriter{},
	}

	pipeline := NewBackupPipeline(config)
	reader := io.NopCloser(bytes.NewReader([]byte("small data")))

	result := pipeline.Execute(context.Background(), reader, "backup.tar")

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	if result.SourceId != "test" {
		t.Errorf("Expected sourceId 'test', got '%s'", result.SourceId)
	}

	if result.SourceName != "Test" {
		t.Errorf("Expected sourceName 'Test', got '%s'", result.SourceName)
	}
}

func TestBackupPipelineSuccessMetadata(t *testing.T) {
	dest := &recordingDestination{}

	config := PipelineConfig{
		SourceId:     "prod-db",
		SourceName:   "Production Database",
		Passphrase:   "secret",
		CompressOnly: false,
		Destinations: []DestinationWriter{dest},
	}

	pipeline := NewBackupPipeline(config)
	reader := io.NopCloser(bytes.NewReader([]byte("database content")))

	result := pipeline.Execute(context.Background(), reader, "db_backup.sql")

	if !result.Success {
		t.Fatalf("Expected success, got error: %v", result.Error)
	}

	_ = result

	if result.Message == "" {
		t.Error("Expected success message")
	}

	if !bytes.Contains([]byte(result.Message), []byte("ms")) {
		t.Errorf("Expected duration in message, got: %s", result.Message)
	}
}
