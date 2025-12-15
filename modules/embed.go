package modules

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed * */*
var embeddedModules embed.FS

var (
	once     sync.Once
	rootPath string
	rootErr  error
)

// Materialize copies the embedded modules directory to a temp directory and returns
// the path to the modules root. It runs once per process.
func Materialize() (string, error) {
	once.Do(func() {
		tmp, err := os.MkdirTemp("", "pltf-modules-*")
		if err != nil {
			rootErr = fmt.Errorf("failed to create temp modules dir: %w", err)
			return
		}

		if err := copyEmbedded(tmp); err != nil {
			rootErr = err
			return
		}

		rootPath = tmp
	})
	return rootPath, rootErr
}

func copyEmbedded(dest string) error {
	return fs.WalkDir(embeddedModules, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".go") {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			return nil
		}

		target := filepath.Join(dest, path)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		f, err := embeddedModules.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, f); err != nil {
			return err
		}
		return nil
	})
}
