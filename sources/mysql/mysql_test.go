package mysql

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

func TestNewMySQLSourceMissingDatabase(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "mysql",
		attrs: map[string]string{},
	}

	_, err := NewMySQLSource(config)
	if err == nil {
		t.Error("Expected error for missing dbname")
	}
}

func TestNewMySQLSourceDefaults(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "mysql",
		attrs: map[string]string{
			"dbname": "testdb",
		},
	}

	src, err := NewMySQLSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", src.host)
	}

	if src.port != 3306 {
		t.Errorf("Expected port 3306, got %d", src.port)
	}

	if src.user != "root" {
		t.Errorf("Expected user 'root', got '%s'", src.user)
	}

	if src.database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", src.database)
	}
}

func TestNewMySQLSourceCustomConfig(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "mysql",
		attrs: map[string]string{
			"dbhost": "db.example.com",
			"dbport": "3307",
			"dbuser": "backup",
			"dbpass": "secret",
			"dbname": "production",
			"options": "--single-transaction --lock-tables=false",
		},
	}

	src, err := NewMySQLSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.host != "db.example.com" {
		t.Errorf("Expected host 'db.example.com', got '%s'", src.host)
	}

	if src.port != 3307 {
		t.Errorf("Expected port 3307, got %d", src.port)
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

	if src.options != "--single-transaction --lock-tables=false" {
		t.Errorf("Expected options '--single-transaction --lock-tables=false', got '%s'", src.options)
	}
}

func TestMySQLInvalidPort(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "mysql",
		attrs: map[string]string{
			"dbname": "testdb",
			"dbport": "invalid",
		},
	}

	src, err := NewMySQLSource(config)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if src.port != 3306 {
		t.Errorf("Expected default port 3306 for invalid port, got %d", src.port)
	}
}
