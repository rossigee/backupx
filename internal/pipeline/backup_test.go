package pipeline

import (
	"bytes"
	"context"
	"io"
	"testing"
)

type mockDestination struct {
	uploadCalled bool
	filename     string
	path         string
}

func (m *mockDestination) UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	m.uploadCalled = true
	m.filename = filename
	m.path = "mock://backup/" + filename
	_, err := io.ReadAll(reader)
	return m.path, err
}

func TestBackupPipelineConfig(t *testing.T) {
	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test Source",
		Passphrase:   "secret",
		CompressOnly: false,
		Destinations: []DestinationWriter{},
	}

	pipeline := NewBackupPipeline(config)

	if pipeline.sourceId != "test" {
		t.Errorf("Expected sourceId 'test', got '%s'", pipeline.sourceId)
	}

	if pipeline.sourceName != "Test Source" {
		t.Errorf("Expected sourceName 'Test Source', got '%s'", pipeline.sourceName)
	}

	if pipeline.passphrase != "secret" {
		t.Errorf("Expected passphrase 'secret', got '%s'", pipeline.passphrase)
	}

	if pipeline.compressOnly {
		t.Error("Expected compressOnly to be false")
	}
}

func TestBackupPipelineNoDestinations(t *testing.T) {
	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test",
		Passphrase:   "secret",
		CompressOnly: true,
		Destinations: []DestinationWriter{},
	}

	pipeline := NewBackupPipeline(config)
	data := []byte("test backup data")
	reader := io.NopCloser(bytes.NewReader(data))

	result := pipeline.Execute(context.Background(), reader, "test.tar")

	if result.SourceId != "test" {
		t.Errorf("Expected sourceId 'test', got '%s'", result.SourceId)
	}

	if !result.Success {
		t.Errorf("Expected success to be true, error: %s", result.Error)
	}

	if !bytes.Contains([]byte(result.BackupPath), []byte("test.tar")) {
		t.Errorf("Expected 'test.tar' in backup path, got %s", result.BackupPath)
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestBackupPipelineWithDestination(t *testing.T) {
	mock := &mockDestination{}
	config := PipelineConfig{
		SourceId:     "test",
		SourceName:   "Test",
		Passphrase:   "secret",
		CompressOnly: true,
		Destinations: []DestinationWriter{mock},
	}

	pipeline := NewBackupPipeline(config)
	data := []byte("test backup data")
	reader := io.NopCloser(bytes.NewReader(data))

	result := pipeline.Execute(context.Background(), reader, "test.tar")

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if !mock.uploadCalled {
		t.Error("Expected destination.UploadReader to be called")
	}

	if mock.filename != "test.tar.gz" {
		t.Errorf("Expected filename 'test.tar.gz', got '%s'", mock.filename)
	}
}
