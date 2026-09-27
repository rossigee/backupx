package postgresql

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

func TestNewPostgreSQLSourceMissingDatabase(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "pgsql",
		attrs: map[string]string{},
	}

	_, err := NewPostgreSQLSource(config)
	if err == nil {
		t.Error("Expected error for missing dbname")
	}
}

func TestNewPostgreSQLSourceDefaults(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "pgsql",
		attrs: map[string]string{
			"dbname": "testdb",
		},
	}

	src, err := NewPostgreSQLSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", src.host)
	}

	if src.port != 5432 {
		t.Errorf("Expected port 5432, got %d", src.port)
	}

	if src.user != "postgres" {
		t.Errorf("Expected user 'postgres', got '%s'", src.user)
	}

	if src.database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", src.database)
	}
}

func TestNewPostgreSQLSourceCustomConfig(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "pgsql",
		attrs: map[string]string{
			"dbhost": "db.example.com",
			"dbport": "5433",
			"dbuser": "backup",
			"dbpass": "secret",
			"dbname": "production",
			"options": "--verbose",
		},
	}

	src, err := NewPostgreSQLSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.host != "db.example.com" {
		t.Errorf("Expected host 'db.example.com', got '%s'", src.host)
	}

	if src.port != 5433 {
		t.Errorf("Expected port 5433, got %d", src.port)
	}

	if src.user != "backup" {
		t.Errorf("Expected user 'backup', got '%s'", src.user)
	}

	if src.password != "secret" {
		t.Errorf("Expected password 'secret', got '%s'", src.password)
	}

	if src.database != "production" {
		t.Errorf("Expected database 'production', got '%s'", src.database)
	}

	if src.options != "--verbose" {
		t.Errorf("Expected options '--verbose', got '%s'", src.options)
	}
}

func TestPostgreSQLInvalidPort(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "pgsql",
		attrs: map[string]string{
			"dbname": "testdb",
			"dbport": "not-a-number",
		},
	}

	src, err := NewPostgreSQLSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.port != 5432 {
		t.Errorf("Expected default port 5432 for invalid port, got %d", src.port)
	}
}
