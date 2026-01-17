package cmd

import (
	"encoding/json"
	"fmt"
	"io"
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

func collectPlanSummaryWithRunner(r *tfDaggerRunner, outDir, planArg, planPath string, stderr io.Writer) (*planSummary, error) {
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
	sum := &planSummary{}

	out, _, err := r.exec([]string{r.engineCmd(), "show", "-json", planArg}, false)
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
			fmt.Fprintf(stderr, "warn: unable to parse plan JSON: %v\n", err)
		}
	} else {
		fmt.Fprintf(stderr, "warn: %s show -json failed: %v\n", r.engineCmd(), err)
	}

	return sum, nil
}
