package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type planSummary struct {
	Added       int
	Changed     int
	Destroyed   int
	Adds        []string
	Changes     []string
	Deletes     []string
	Text        string
	RawPlanArgs []string
}

type tfPlanJSON struct {
	ResourceChanges []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

func collectPlanSummary(outDir, planFile string) (*planSummary, error) {
	if strings.TrimSpace(planFile) == "" {
		return nil, nil
	}
	planPath := planFile
	if !filepath.IsAbs(planFile) {
		planPath = filepath.Clean(planPath)
		if !strings.HasPrefix(planPath, outDir) {
			planPath = filepath.Join(outDir, planPath)
		}
	}
	if _, err := os.Stat(planPath); err != nil {
		return nil, err
	}
	sum := &planSummary{}

	out, err := runCmdOutput(outDir, "terraform", "show", "-json", planPath)
	if err == nil {
		var plan tfPlanJSON
		if err := json.Unmarshal([]byte(out), &plan); err == nil {
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
			fmt.Fprintf(os.Stderr, "warn: unable to parse plan JSON: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "warn: terraform show -json failed: %v\n", err)
	}

	if text, err := runCmdOutput(outDir, "terraform", "show", "-no-color", planPath); err == nil {
		// Keep full plan output; GitHub comment limit is large enough for typical plans.
		sum.Text = strings.TrimSpace(text)
	}

	return sum, nil
}
