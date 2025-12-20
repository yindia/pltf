package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"pltf/pkg/config"
)

var (
	applyFile        string
	applyEnv         string
	applyOut         string
	applyModulesDir  string
	applyVars        []string
	applyTargets     []string
	applyParallel    int
	applyLock        bool
	applyLockTime    string
	applyNoColor     bool
	applyInput       bool
	applyRefresh     bool
	applyAutoApprove bool

	destroyFile        string
	destroyEnv         string
	destroyOut         string
	destroyModulesDir  string
	destroyVars        []string
	destroyTargets     []string
	destroyParallel    int
	destroyLock        bool
	destroyLockTime    string
	destroyNoColor     bool
	destroyInput       bool
	destroyRefresh     bool
	destroyAutoApprove bool

	planFile       string
	planEnv        string
	planOut        string
	planModulesDir string
	planVars       []string
	planTargets    []string
	planParallel   int
	planLock       bool
	planLockTime   string
	planNoColor    bool
	planInput      bool
	planRefresh    bool
	planDetailed   bool
	planOutFile    string

	outputFile       string
	outputEnv        string
	outputOut        string
	outputModulesDir string
	outputVar        string
	outputJSON       bool
	outputNoColor    bool

	unlockFile       string
	unlockEnv        string
	unlockOut        string
	unlockModulesDir string
	unlockLockID     string
	unlockNoColor    bool
	unlockLock       bool
	unlockLockTime   string

	graphFile       string
	graphEnv        string
	graphOut        string
	graphModulesDir string
	graphVars       []string
	graphMode       string
	graphOutFile    string
	graphPlanFile   string
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Args:  cobra.NoArgs,
	Short: "Generate and apply Terraform for a spec",
	Long: `Render Terraform from an Environment or Service spec, ensure the backend bucket,
then run 'terraform apply'. Supports Terraform-style flags like targets, lock timeout,
parallelism, refresh control, and color toggles. Defaults to embedded modules and the
standard output layout unless overridden.`,
	Example: `  pltf terraform apply -f env.yaml -e prod
  pltf terraform apply -f service.yaml -e dev -m ./modules -o ./.pltf/service/payments/dev --target=module.eks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTfWithAction("apply", applyFile, applyEnv, applyModulesDir, applyOut, applyVars, "", tfExecOpts{
			targets:      applyTargets,
			parallelism:  applyParallel,
			lock:         applyLock,
			lockTimeout:  applyLockTime,
			noColor:      applyNoColor,
			input:        applyInput,
			refresh:      &applyRefresh,
			jsonOutput:   false,
			planFile:     "",
			detailedExit: false,
			autoApprove:  applyAutoApprove,
		})
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Args:  cobra.NoArgs,
	Short: "Generate a DOT graph for a spec (terraform graph or spec dependency graph)",
	Long: `Render Terraform (if needed) and produce a DOT graph. By default runs 'terraform graph'
against the generated stack. With --mode=spec, emits a dependency graph from the env/service
YAML (links and module references) without invoking Terraform.`,
	Example: `  pltf terraform graph -f env.yaml -e dev > graph.dot
  pltf terraform graph -f service.yaml -e dev --mode=spec --out-file=spec.dot`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGraph(graphMode, graphFile, graphEnv, graphModulesDir, graphOut, graphVars, graphOutFile, graphPlanFile)
	},
}

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Args:  cobra.NoArgs,
	Short: "Generate (if needed) and destroy Terraform for a spec",
	Long: `Render Terraform if missing, then run 'terraform destroy'. Mirrors apply
defaults (modules, output layout) and exposes Terraform knobs for targets, locking,
refresh behavior, and color.`,
	Example: `  pltf terraform destroy -f env.yaml -e prod
  pltf terraform destroy -f service.yaml -e dev --target=module.app-bucket`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTfWithAction("destroy", destroyFile, destroyEnv, destroyModulesDir, destroyOut, destroyVars, "", tfExecOpts{
			targets:     destroyTargets,
			parallelism: destroyParallel,
			lock:        destroyLock,
			lockTimeout: destroyLockTime,
			noColor:     destroyNoColor,
			input:       destroyInput,
			refresh:     &destroyRefresh,
			autoApprove: destroyAutoApprove,
		})
	},
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Args:  cobra.NoArgs,
	Short: "Generate (if needed) and run terraform plan for a spec",
	Long: `Render Terraform and run 'terraform plan'. Supports detailed exit codes,
plan file output, targets, locking, refresh toggles, and parallelism. Ideal for CI or
local dry runs with the same generation defaults as apply.`,
	Example: `  pltf terraform plan -f env.yaml -e prod
  pltf terraform plan -f service.yaml -e dev --detailed-exitcode --plan-file=/tmp/plan.tfplan`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTfWithAction("plan", planFile, planEnv, planModulesDir, planOut, planVars, "", tfExecOpts{
			targets:      planTargets,
			parallelism:  planParallel,
			lock:         planLock,
			lockTimeout:  planLockTime,
			noColor:      planNoColor,
			input:        planInput,
			refresh:      &planRefresh,
			planFile:     planOutFile,
			detailedExit: planDetailed,
		})
	},
}

var outputCmd = &cobra.Command{
	Use:   "output",
	Args:  cobra.NoArgs,
	Short: "Show terraform outputs for a generated spec",
	Long:  "Print Terraform outputs for the rendered stack. Supports JSON output for scripting and color toggles.",
	Example: `  pltf terraform output -f env.yaml -e prod
  pltf terraform output -f service.yaml -e dev --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTfWithAction("output", outputFile, outputEnv, outputModulesDir, outputOut, nil, outputVar, tfExecOpts{
			noColor:    outputNoColor,
			jsonOutput: outputJSON,
		})
	},
}

var unlockCmd = &cobra.Command{
	Use:   "force-unlock",
	Args:  cobra.NoArgs,
	Short: "Force unlock Terraform state for a spec",
	Long:  "Run 'terraform force-unlock' against the generated stack. Use only to clear stale locks after verifying no active operation.",
	Example: `  pltf terraform force-unlock -f env.yaml -e prod --lock-id=12345
  pltf terraform force-unlock -f service.yaml -e dev --lock-id=$(cat .terraform.tfstate.lock.info | jq -r .ID)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(unlockLockID) == "" {
			return fmt.Errorf("--lock-id is required")
		}
		return runTfWithAction("force-unlock", unlockFile, unlockEnv, unlockModulesDir, unlockOut, nil, unlockLockID, tfExecOpts{
			noColor:     unlockNoColor,
			lock:        unlockLock,
			lockTimeout: unlockLockTime,
		})
	},
}

type tfExecOpts struct {
	targets      []string
	parallelism  int
	lock         bool
	lockTimeout  string
	noColor      bool
	input        bool
	refresh      *bool
	planFile     string
	detailedExit bool
	jsonOutput   bool
	autoApprove  bool
}

type stackContext struct {
	kind   string
	env    string
	envCfg *config.EnvironmentConfig
	outDir string
}

func prepareStackContext(file, env, out string) (stackContext, error) {
	var ctx stackContext

	kind, err := config.DetectKind(defaultString(file, "env.yaml"))
	if err != nil {
		return ctx, err
	}
	ctx.kind = kind

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(defaultString(file, "env.yaml"))
		if err != nil {
			return ctx, err
		}
		env, err = selectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return ctx, err
		}
		ctx.envCfg = envCfg
		if out == "" {
			ctx.outDir = filepath.Join(".pltf", envCfg.Metadata.Name, "env", env)
		} else {
			ctx.outDir = out
		}
	case "Service":
		svcCfg, envCfg, err := config.LoadService(defaultString(file, "service.yaml"))
		if err != nil {
			return ctx, err
		}
		env, err = selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return ctx, err
		}
		ctx.envCfg = envCfg
		if out == "" {
			ctx.outDir = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "env", env)
		} else {
			ctx.outDir = out
		}
	default:
		return ctx, fmt.Errorf("unknown kind %q", kind)
	}

	ctx.env = env
	ctx.outDir = filepath.Clean(ctx.outDir)
	return ctx, nil
}

func runTfWithAction(action, file, env, modules, out string, vars []string, lockID string, opts tfExecOpts) error {
	// Generate configs first
	if err := autoGenerate(file, env, modules, out, vars); err != nil {
		return err
	}

	ctx, err := prepareStackContext(file, env, out)
	if err != nil {
		return err
	}

	bk, err := computeBackend(ctx.envCfg, ctx.env)
	if err != nil {
		return err
	}
	if isS3Backend(bk) {
		if err := ensureS3Bucket(bk.bucket, bk.region); err != nil {
			return fmt.Errorf("failed to ensure backend bucket: %w", err)
		}
	}

	if err := runCmd(ctx.outDir, "terraform", "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	common := func(args []string) []string {
		args = appendTfCommonArgs(args, opts)
		return args
	}

	var runErr error
	var planSum *planSummary
	runStatus := "succeeded"
	var planArgs []string
	var planExit int
	switch action {
	case "apply":
		args := []string{"apply"}
		if opts.autoApprove {
			args = append(args, "-auto-approve")
		}
		if err := runCmd(ctx.outDir, "terraform", common(args)...); err != nil {
			runErr = fmt.Errorf("terraform apply failed: %w", err)
		}
	case "destroy":
		args := []string{"destroy"}
		if opts.autoApprove {
			args = append(args, "-auto-approve")
		}
		if err := runCmd(ctx.outDir, "terraform", common(args)...); err != nil {
			runErr = fmt.Errorf("terraform destroy failed: %w", err)
		}
	case "plan":
		args := []string{"plan"}
		if opts.detailedExit {
			args = append(args, "-detailed-exitcode")
		}
		planPath := opts.planFile
		planArg := opts.planFile
		tempPlan := false
		if strings.TrimSpace(planPath) == "" {
			planArg = ".pltf-plan.tfplan"
			planPath = filepath.Join(ctx.outDir, planArg)
			tempPlan = true
		} else {
			if filepath.IsAbs(planPath) {
				planArg = planPath
			} else {
				planArg = planPath
				planPath = filepath.Join(ctx.outDir, planPath)
			}
		}
		args = append(args, "-out="+planArg)
		planArgs = append(planArgs, common(args)...)
		planExit, runErr = runCmdExit(ctx.outDir, "terraform", planArgs...)
		if runErr != nil && !(opts.detailedExit && planExit == 2) {
			runErr = fmt.Errorf("terraform plan failed: %w", runErr)
		}
		if opts.detailedExit && planExit == 2 && runErr == nil {
			runStatus = "changes"
		}
		planPathOnDisk := planPath
		if !filepath.IsAbs(planPathOnDisk) {
			planPathOnDisk = filepath.Clean(filepath.Join(ctx.outDir, planPathOnDisk))
		}
		if sum, err := collectPlanSummary(ctx.outDir, planPathOnDisk); err == nil {
			planSum = sum
			planSum.RawPlanArgs = planArgs
		} else {
			fmt.Fprintf(os.Stderr, "warn: failed to collect plan summary: %v\n", err)
		}
		if tempPlan {
			_ = os.Remove(planPathOnDisk)
		}
	case "output":
		args := []string{"output"}
		if lockID != "" {
			args = append(args, lockID)
		}
		if opts.jsonOutput {
			args = append(args, "-json")
		}
		if err := runCmd(ctx.outDir, "terraform", common(args)...); err != nil {
			runErr = fmt.Errorf("terraform output failed: %w", err)
		}
	case "force-unlock":
		args := []string{"force-unlock", "-force", lockID}
		if err := runCmd(ctx.outDir, "terraform", common(args)...); err != nil {
			runErr = fmt.Errorf("terraform force-unlock failed: %w", err)
		}
	}

	if action == "plan" || action == "apply" {
		status := tfRunSummary{
			Action: action,
			Spec:   file,
			Env:    env,
			OutDir: ctx.outDir,
			Plan:   planSum,
		}
		if runErr != nil {
			status.Status = "failed"
			status.Err = runErr.Error()
		} else {
			status.Status = runStatus
		}
		if status.Plan != nil {
			status.AI = maybeAICritique(status)
		}
		if err := maybeUpsertPRComment(status); err != nil {
			fmt.Fprintf(os.Stderr, "warn: failed to update PR comment: %v\n", err)
		}
	}

	return runErr
}

func init() {
	terraformCmd := &cobra.Command{Use: "terraform", Short: "Terraform helpers (generate+init+tf commands)"}
	rootCmd.AddCommand(terraformCmd)
	terraformCmd.AddCommand(applyCmd)
	terraformCmd.AddCommand(destroyCmd)
	terraformCmd.AddCommand(planCmd)
	terraformCmd.AddCommand(outputCmd)
	terraformCmd.AddCommand(unlockCmd)
	terraformCmd.AddCommand(graphCmd)

	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	applyCmd.Flags().StringVarP(&applyEnv, "env", "e", "", "Environment key to render (dev, prod, etc.)")
	applyCmd.Flags().StringVarP(&applyModulesDir, "modules", "m", "", "Override modules root; defaults to embedded modules")
	applyCmd.Flags().StringVarP(&applyOut, "out", "o", "", "Output directory for generated Terraform")
	applyCmd.Flags().StringArrayVarP(&applyVars, "var", "v", nil, "Override variable as key=value; merges over vars and supports bool/int/JSON/list parsing. Can be repeated for multiple overrides.")
	applyCmd.Flags().StringArrayVarP(&applyTargets, "target", "t", nil, "Optional Terraform target address (repeatable)")
	applyCmd.Flags().IntVarP(&applyParallel, "parallelism", "p", 0, "Limit Terraform parallelism (0 = default)")
	applyCmd.Flags().BoolVarP(&applyLock, "lock", "l", true, "Lock state when locking is supported")
	applyCmd.Flags().StringVarP(&applyLockTime, "lock-timeout", "T", "", "Lock timeout (e.g. 0s, 30s)")
	applyCmd.Flags().BoolVarP(&applyNoColor, "no-color", "C", false, "Disable color output")
	applyCmd.Flags().BoolVarP(&applyInput, "input", "i", false, "Ask for input if necessary (default false)")
	applyCmd.Flags().BoolVarP(&applyRefresh, "refresh", "r", true, "Update state prior to actions")
	applyCmd.Flags().BoolVar(&applyAutoApprove, "auto-approve", false, "Pass -auto-approve to terraform apply")

	destroyCmd.Flags().StringVarP(&destroyFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	destroyCmd.Flags().StringVarP(&destroyEnv, "env", "e", "", "Environment key to render (dev, prod, etc.)")
	destroyCmd.Flags().StringVarP(&destroyModulesDir, "modules", "m", "", "Override modules root; defaults to embedded modules")
	destroyCmd.Flags().StringVarP(&destroyOut, "out", "o", "", "Output directory for generated Terraform")
	destroyCmd.Flags().StringArrayVar(&destroyVars, "var", nil, "Override variable as key=value; merges over vars and supports bool/int/JSON/list parsing. Can be repeated for multiple overrides.")
	destroyCmd.Flags().StringArrayVarP(&destroyTargets, "target", "t", nil, "Optional Terraform target address (repeatable)")
	destroyCmd.Flags().IntVarP(&destroyParallel, "parallelism", "p", 0, "Limit Terraform parallelism (0 = default)")
	destroyCmd.Flags().BoolVarP(&destroyLock, "lock", "l", true, "Lock state when locking is supported")
	destroyCmd.Flags().StringVarP(&destroyLockTime, "lock-timeout", "T", "", "Lock timeout (e.g. 0s, 30s)")
	destroyCmd.Flags().BoolVarP(&destroyNoColor, "no-color", "C", false, "Disable color output")
	destroyCmd.Flags().BoolVarP(&destroyInput, "input", "i", false, "Ask for input if necessary (default false)")
	destroyCmd.Flags().BoolVarP(&destroyRefresh, "refresh", "r", true, "Update state prior to actions")
	destroyCmd.Flags().BoolVar(&destroyAutoApprove, "auto-approve", false, "Pass -auto-approve to terraform destroy")

	planCmd.Flags().StringVarP(&planFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	planCmd.Flags().StringVarP(&planEnv, "env", "e", "", "Environment key to render (dev, prod, etc.)")
	planCmd.Flags().StringVarP(&planModulesDir, "modules", "m", "", "Override modules root; defaults to embedded modules")
	planCmd.Flags().StringVarP(&planOut, "out", "o", "", "Output directory for generated Terraform")
	planCmd.Flags().StringArrayVarP(&planVars, "var", "v", nil, "Override variable as key=value; merges over vars and supports bool/int/JSON/list parsing. Can be repeated for multiple overrides.")
	planCmd.Flags().StringArrayVarP(&planTargets, "target", "t", nil, "Optional Terraform target address (repeatable)")
	planCmd.Flags().IntVarP(&planParallel, "parallelism", "p", 0, "Limit Terraform parallelism (0 = default)")
	planCmd.Flags().BoolVarP(&planLock, "lock", "l", true, "Lock state when locking is supported")
	planCmd.Flags().StringVarP(&planLockTime, "lock-timeout", "T", "", "Lock timeout (e.g. 0s, 30s)")
	planCmd.Flags().BoolVarP(&planNoColor, "no-color", "C", false, "Disable color output")
	planCmd.Flags().BoolVarP(&planInput, "input", "i", false, "Ask for input if necessary (default false)")
	planCmd.Flags().BoolVarP(&planRefresh, "refresh", "r", true, "Update state prior to actions")
	planCmd.Flags().BoolVarP(&planDetailed, "detailed-exitcode", "d", false, "Use detailed exit codes for plan (2 = changes present)")
	planCmd.Flags().StringVarP(&planOutFile, "plan-file", "P", "", "Write plan to a file (terraform -out)")

	outputCmd.Flags().StringVarP(&outputFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	outputCmd.Flags().StringVarP(&outputEnv, "env", "e", "", "Environment key to render (dev, prod, etc.)")
	outputCmd.Flags().StringVarP(&outputModulesDir, "modules", "m", "", "Override modules root; defaults to embedded modules")
	outputCmd.Flags().StringVarP(&outputOut, "out", "o", "", "Output directory for generated Terraform")
	outputCmd.Flags().StringVarP(&outputVar, "var", "v", "", "Specific output name to show (optional)")
	outputCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Render output as JSON")
	outputCmd.Flags().BoolVarP(&outputNoColor, "no-color", "C", false, "Disable color output")

	unlockCmd.Flags().StringVarP(&unlockFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	unlockCmd.Flags().StringVarP(&unlockEnv, "env", "e", "", "Environment key to render (dev, prod, etc.)")
	unlockCmd.Flags().StringVarP(&unlockModulesDir, "modules", "m", "", "Override modules root; defaults to embedded modules")
	unlockCmd.Flags().StringVarP(&unlockOut, "out", "o", "", "Output directory for generated Terraform")
	unlockCmd.Flags().StringVar(&unlockLockID, "lock-id", "", "Terraform lock ID to unlock")
	unlockCmd.MarkFlagRequired("lock-id")
	unlockCmd.Flags().BoolVarP(&unlockNoColor, "no-color", "C", false, "Disable color output")
	unlockCmd.Flags().BoolVarP(&unlockLock, "lock", "l", true, "Lock state when locking is supported")
	unlockCmd.Flags().StringVarP(&unlockLockTime, "lock-timeout", "T", "", "Lock timeout (e.g. 0s, 30s)")

	graphCmd.Flags().StringVarP(&graphFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	graphCmd.Flags().StringVarP(&graphEnv, "env", "e", "", "Environment key to render (dev, prod, etc.)")
	graphCmd.Flags().StringVarP(&graphModulesDir, "modules", "m", "", "Override modules root; defaults to embedded modules")
	graphCmd.Flags().StringVarP(&graphOut, "out", "o", "", "Output directory for generated Terraform (for terraform mode)")
	graphCmd.Flags().StringArrayVarP(&graphVars, "var", "v", nil, "Override variable as key=value; merges over vars and supports bool/int/JSON/list parsing. Used for terraform mode generation.")
	graphCmd.Flags().StringVarP(&graphMode, "mode", "", "terraform", "Graph mode: terraform (runs 'terraform graph') or spec (builds module dependency graph from YAML)")
	graphCmd.Flags().StringVarP(&graphOutFile, "out-file", "", "", "Write DOT output to a file instead of stdout")
	graphCmd.Flags().StringVarP(&graphPlanFile, "plan-file", "P", "", "Use an existing plan file for terraform graph (passed as -plan=...)")
}
