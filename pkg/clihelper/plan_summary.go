package clihelper

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PlanSummary struct {
	Added       int
	Changed     int
	Destroyed   int
	Adds        []string
	Changes     []string
	Deletes     []string
	RawPlanArgs []string
	PlanJSON    string
}

type tfPlanJSON struct {
	ResourceChanges []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

func CollectPlanSummaryWithRunner(outDir, planPath, planJSON string, stderr io.Writer) (*PlanSummary, error) {
	if strings.TrimSpace(planPath) == "" {
		return nil, nil
	}
	planPathOnDisk := planPath
	if !filepath.IsAbs(planPathOnDisk) {
		planPathOnDisk = filepath.Clean(planPathOnDisk)
		if !strings.HasPrefix(planPathOnDisk, outDir) {
			planPathOnDisk = filepath.Join(outDir, planPathOnDisk)
		}
	}
	if _, err := os.Stat(planPathOnDisk); err != nil {
		return nil, err
	}
	sum := &PlanSummary{}

	if strings.TrimSpace(planJSON) == "" {
		return sum, nil
	}
	var plan tfPlanJSON
	if err := json.Unmarshal([]byte(planJSON), &plan); err == nil {
		for _, rc := range plan.ResourceChanges {
			actions := map[string]bool{}
			for _, a := range rc.Change.Actions {
				actions[a] = true
			}
			switch {
			case actions["create"] && actions["delete"]:
				sum.Changed++
				sum.Changes = append(sum.Changes, rc.Address)
			case actions["update"]:
				sum.Changed++
				sum.Changes = append(sum.Changes, rc.Address)
			case actions["create"]:
				sum.Added++
				sum.Adds = append(sum.Adds, rc.Address)
			case actions["delete"]:
				sum.Destroyed++
				sum.Deletes = append(sum.Deletes, rc.Address)
			}
		}
	} else {
		fmt.Fprintf(stderr, "warn: unable to parse plan JSON: %v\n", err)
	}

	return sum, nil
}
