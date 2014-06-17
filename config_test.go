package main

import "testing"

func TestConfigFiles(t *testing.T) {
	filename := "some/file.yml"
	options := Options{config: filename}
	files := configFiles(options)
	if len(files) > 1 {
		t.Errorf("Config files should be just [%s], got %v", filename, files)
	}

	files = configFiles(Options{})
	if len(files) != 4 {
		t.Errorf("Config files should be [crane.json, crane.yaml, crane.yml, Cranefile], got %v", files)
	}
}
