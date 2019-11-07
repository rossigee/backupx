package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type conf struct {
	Spec struct {
		Source struct {
			Name string `yaml:"name"`
			Description string `yaml:"description"`
		} `yaml:"source"`
		Destinations []struct {
			Name string `yaml:"name"`
			Description string `yaml:"description"`
		} `yaml:"destinations"`
		Notifiers []struct {
			Name string `yaml:"name"`
			Description string `yaml:"description"`
		} `yaml:"notifiers"`
	} `yaml:"spec"`
}

func (c *conf) parseYamlConf(filename string) *conf {
	rawYaml, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v ", err)
	}
	err = yaml.Unmarshal(rawYaml, c)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v ", err)
	}
	return c
}
