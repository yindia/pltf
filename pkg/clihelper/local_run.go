package clihelper

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type localRunMeta struct {
	ID         string `json:"id"`
	Action     string `json:"action"`
	Spec       string `json:"spec"`
	Env        string `json:"env"`
	OutDir     string `json:"out_dir"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
}

const localRunRetention = 50

func StartLocalRun(action, spec, env, outDir string) (string, func(error)) {
	meta := localRunMeta{
		ID:        newLocalRunID(),
		Action:    action,
		Spec:      spec,
		Env:       env,
		OutDir:    outDir,
		Status:    "running",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	path := localRunPath(meta.ID)
	_ = writeLocalRunMeta(path, meta)

	return meta.ID, func(err error) {
		if err != nil {
			meta.Status = "failed"
			meta.Error = err.Error()
		} else {
			meta.Status = "succeeded"
		}
		meta.FinishedAt = time.Now().Format(time.RFC3339)
		_ = writeLocalRunMeta(path, meta)
	}
}

func newLocalRunID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("run-%d", time.Now().UnixNano())
}

func localRunPath(id string) string {
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}
	return filepath.Join(root, ".pltf", "runs", id+".json")
}

func writeLocalRunMeta(path string, meta localRunMeta) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	pruneLocalRuns(filepath.Dir(path), localRunRetention)
	return nil
}

func pruneLocalRuns(dir string, keep int) {
	if keep <= 0 {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type fileInfo struct {
		name string
		mod  time.Time
	}
	var files []fileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{name: entry.Name(), mod: info.ModTime()})
	}
	if len(files) <= keep {
		return
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].mod.After(files[j].mod)
	})
	for _, f := range files[keep:] {
		_ = os.Remove(filepath.Join(dir, f.name))
	}
}
