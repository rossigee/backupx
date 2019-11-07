package main

import (
	"testing"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func TestYamlParser(t *testing.T) {
	var c conf
	rawYaml, err := ioutil.ReadFile("conf_test.yaml")
	if err != nil {
		t.Errorf("Error reading YAML file: %v ", err)
	}
	err = yaml.Unmarshal(rawYaml, &c)
	if err != nil {
		t.Errorf("Error parsing YAML file: %v ", err)
	}
	if c.Spec.Source.Name != "testsource" {
		t.Errorf("Source name not as expected: %s", c.Spec.Source.Name)
	}
	if len(c.Spec.Destinations) != 2 {
		t.Errorf("Unexpected Destination count: %d", len(c.Spec.Destinations))
	}
	if len(c.Spec.Notifiers) != 2 {
		t.Errorf("Unexpected Notifier count: %d", len(c.Spec.Notifiers))
	}
}
