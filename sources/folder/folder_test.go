package folder

import (
	"testing"
)

type testConfig struct {
	id    string
	typ   string
	attrs map[string]string
}

func (c testConfig) GetId() string {
	return c.id
}

func (c testConfig) GetType() string {
	return c.typ
}

func (c testConfig) GetName() string {
	return c.id
}

func (c testConfig) GetOtherAttributes() map[string]string {
	return c.attrs
}

func TestNewFolderSourceMissingPath(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "folder",
		attrs: map[string]string{},
	}

	_, err := NewFolderSource(config)
	if err == nil {
		t.Error("Expected error for missing path")
	}
}

func TestNewFolderSourceDefaults(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "folder",
		attrs: map[string]string{
			"path": "/var/data",
		},
	}

	src, err := NewFolderSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.path != "/var/data" {
		t.Errorf("Expected path '/var/data', got '%s'", src.path)
	}

	if len(src.excludes) != 0 {
		t.Errorf("Expected no excludes, got %d", len(src.excludes))
	}
}

func TestNewFolderSourceWithExcludes(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "folder",
		attrs: map[string]string{
			"path":     "/var/data",
			"excludes": "*.log, /tmp, .cache",
		},
	}

	src, err := NewFolderSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if len(src.excludes) != 3 {
		t.Errorf("Expected 3 excludes, got %d", len(src.excludes))
	}

	expected := []string{"*.log", "/tmp", ".cache"}
	for i, ex := range src.excludes {
		if ex != expected[i] {
			t.Errorf("Expected exclude '%s', got '%s'", expected[i], ex)
		}
	}
}
