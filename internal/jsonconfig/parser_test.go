package jsonconfig

import (
	"testing"

	"github.com/rossigee/backupx/config"
)

func TestJSONParser(t *testing.T) {
	var i config.IBackupConfig

	i, err := ParseJSONConfigFile("testdata/test1.json")
	if err != nil {
		t.Errorf("Error reading JSON file: %v ", err)
	}

	// TODO: Move this to a common area to share w/YAML parser test
	if i.GetId() != "test1" {
		t.Errorf("Backup spec id not as expected: %s", i.GetId())
	}
	if i.GetName() != "Test backup #1" {
		t.Errorf("Backup spec name not as expected: %s", i.GetName())
	}
	var sources = i.GetSourceConfigs()
	if len(sources) != 1 {
		t.Errorf("Unexpected Source count: %d", len(sources))
	} else {
		var source = sources[0]
		if source.GetName() != "Test data volume 1" {
			t.Errorf("Source name not as expected: %s", source.GetName())
		}
	}
	var destinations = i.GetDestinationConfigs()
	if len(destinations) != 1 {
		t.Errorf("Unexpected Destination count: %d", len(destinations))
	}
	var notifications = i.GetNotificationConfigs()
	if len(notifications) != 1 {
		t.Errorf("Unexpected Notification count: %d", len(notifications))
	}
}

func TestJSONParserInvalidJSON(t *testing.T) {
	// TODO: stub ugly error output?
	_, err := ParseJSONConfigFile("testdata/test_invalid_json.json")
	if err == nil {
		t.Errorf("Didn't fail to read invalid JSON file.")
	}
}

func TestJSONParserFileNotFound(t *testing.T) {
	// TODO: stub ugly error output?
	_, err := ParseJSONConfigFile("testdata/test_doesnt_exist.json")
	if err == nil {
		t.Errorf("Didn't fail to read non-existent file.")
	}
}
