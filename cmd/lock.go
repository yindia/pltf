package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func requiresEnvLock(action string) bool {
	switch action {
	case "apply", "destroy":
		return true
	default:
		return false
	}
}

func acquireEnvLock(envName, envKey string) (func(), error) {
	if envName == "" || envKey == "" {
		return func() {}, fmt.Errorf("env lock requires env name and env key")
	}
	root, err := os.Getwd()
	if err != nil {
		return func() {}, fmt.Errorf("resolve working directory: %w", err)
	}
	lockDir := filepath.Join(root, ".pltf", envName, "locks")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return func() {}, fmt.Errorf("create lock directory: %w", err)
	}
	lockPath := filepath.Join(lockDir, fmt.Sprintf("%s.lock", envKey))
	f, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return func() {}, fmt.Errorf("environment %q is locked (lock file %s exists)", envKey, lockPath)
	}
	_, _ = fmt.Fprintf(f, "pid=%d\nstarted=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	_ = f.Close()

	release := func() {
		_ = os.Remove(lockPath)
	}
	return release, nil
}
