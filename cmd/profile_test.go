package cmd

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadProfileFromEnv(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	prof := profileConfig{ModulesRoot: dir, DefaultEnv: "dev"}
	data, err := yaml.Marshal(prof)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(profilePath, data, 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	os.Setenv("PLTF_PROFILE", profilePath)
	defer os.Unsetenv("PLTF_PROFILE")

	profileOnce = sync.Once{}
	profileData = nil
	profileErr = nil

	p := loadProfile()
	if p == nil {
		t.Fatalf("expected profile loaded")
	}
	if p.DefaultEnv != "dev" {
		t.Fatalf("expected default env dev, got %s", p.DefaultEnv)
	}
}
