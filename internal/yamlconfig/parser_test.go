package yamlconfig

import (
	"testing"

	"github.com/rossigee/backupx/config"
)

func TestYAMLParser(t *testing.T) {
	var i config.IBackupConfig

	i, err := ParseYAMLConfigFile("testdata/test1.yaml")
	if err != nil {
		t.Errorf("Error reading YAML file: %v ", err)
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

func TestYAMLParserInvalidYAML(t *testing.T) {
	// TODO: stub ugly error output?
	_, err := ParseYAMLConfigFile("testdata/test_invalid_yaml.yaml")
	if err == nil {
		t.Errorf("Didn't fail to read invalid YAML file.")
	}
}

func TestYAMLParserFileNotFound(t *testing.T) {
	// TODO: stub ugly error output?
	_, err := ParseYAMLConfigFile("testdata/conf_test_doesnt_exist.yaml")
	if err == nil {
		t.Errorf("Didn't fail to read non-existent file.")
	}
}
