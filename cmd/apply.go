package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"dagger.io/dagger"
	tfscanner "github.com/aquasecurity/defsec/pkg/scanners/terraform"
	defsecTypes "github.com/aquasecurity/defsec/pkg/types"
	"github.com/spf13/cobra"

	"pltf/pkg/backend"
	"pltf/pkg/clihelper"
	"pltf/pkg/config"
	"pltf/pkg/daggerx"
	"pltf/pkg/generate"
	terraform "pltf/pkg/terraform"
	rover "rover"
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
	applyAutoApprove = true

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
	destroyAutoApprove = true

	planFile         string
	planEnv          string
	planOut          string
	planModulesDir   string
	planVars         []string
	planTargets      []string
	planParallel     int
	planLock         bool
	planLockTime     string
	planNoColor      bool
	planInput        bool
	planRefresh      bool
	planDetailed     bool
	planOutFile      string
	planRover        bool
	planScan         bool
	planCost         bool
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

const defaultTfEngine = "terraform"

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
		return runTfWithAction("apply", applyFile, applyEnv, applyModulesDir, applyOut, applyVars, "", clihelper.TerraformExecOptions{
			Targets:      applyTargets,
			Parallelism:  applyParallel,
			Lock:         applyLock,
			LockTimeout:  applyLockTime,
			NoColor:      applyNoColor,
			Input:        applyInput,
			Refresh:      &applyRefresh,
			JSONOutput:   false,
			PlanFile:     "",
			DetailedExit: false,
			AutoApprove:  applyAutoApprove,
		})
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Args:  cobra.NoArgs,
	Short: "Generate a graph for a spec (terraform or spec)",
	Long: `Render Terraform (if needed) and produce a graph. By default runs 'terraform graph'
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
		return runTfWithAction("destroy", destroyFile, destroyEnv, destroyModulesDir, destroyOut, destroyVars, "", clihelper.TerraformExecOptions{
			Targets:     destroyTargets,
			Parallelism: destroyParallel,
			Lock:        destroyLock,
			LockTimeout: destroyLockTime,
			NoColor:     destroyNoColor,
			Input:       destroyInput,
			Refresh:     &destroyRefresh,
			AutoApprove: destroyAutoApprove,
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
  pltf terraform plan -f service.yaml -e dev --detailed-exitcode --plan-file=/tmp/plan.tfplan
  pltf terraform plan -f env.yaml -e prod --rover   # renders plan.json and opens rover (https://github.com/yindia/rover)
  pltf terraform plan -f env.yaml -e prod --scan    # run tfsec against generated TF
  pltf terraform plan -f env.yaml -e prod --cost    # run infracost breakdown (if infracost binary present)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTfWithAction("plan", planFile, planEnv, planModulesDir, planOut, planVars, "", clihelper.TerraformExecOptions{
			Targets:      planTargets,
			Parallelism:  planParallel,
			Lock:         planLock,
			LockTimeout:  planLockTime,
			NoColor:      planNoColor,
			Input:        planInput,
			Refresh:      &planRefresh,
			PlanFile:     planOutFile,
			DetailedExit: planDetailed,
			Rover:        planRover,
			Scan:         planScan,
			Cost:         planCost,
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
		return runTfWithAction("output", outputFile, outputEnv, outputModulesDir, outputOut, nil, outputVar, clihelper.TerraformExecOptions{
			NoColor:    outputNoColor,
			JSONOutput: outputJSON,
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
		return runTfWithAction("force-unlock", unlockFile, unlockEnv, unlockModulesDir, unlockOut, nil, unlockLockID, clihelper.TerraformExecOptions{
			NoColor:     unlockNoColor,
			Lock:        unlockLock,
			LockTimeout: unlockLockTime,
		})
	},
}

type stackContext struct {
	kind   string
	env    string
	envCfg *config.EnvironmentConfig
	svcCfg *config.ServiceConfig
	outDir string
}

func hasImageBuilds(ctx stackContext) bool {
	if ctx.envCfg != nil && len(ctx.envCfg.Images) > 0 {
		return true
	}
	if ctx.svcCfg != nil && len(ctx.svcCfg.Images) > 0 {
		return true
	}
	return false
}

func prepareWorkspaceContext(file, env, out string) (stackContext, error) {
	var ctx stackContext

	kind, err := config.DetectKind(clihelper.DefaultString(file, "env.yaml"))
	if err != nil {
		return ctx, err
	}
	ctx.kind = kind

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(clihelper.DefaultString(file, "env.yaml"))
		if err != nil {
			return ctx, err
		}
		env, err = clihelper.SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return ctx, err
		}
		ctx.envCfg = envCfg
		ctx.svcCfg = nil
		if out == "" {
			ctx.outDir = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		} else {
			ctx.outDir = out
		}
	case "Service":
		svcCfg, envCfg, err := config.LoadService(clihelper.DefaultString(file, "service.yaml"))
		if err != nil {
			return ctx, err
		}
		env, err = clihelper.SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return ctx, err
		}
		ctx.envCfg = envCfg
		ctx.svcCfg = svcCfg
		if out == "" {
			ctx.outDir = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		} else {
			ctx.outDir = out
		}
	case "Stack":
		return ctx, fmt.Errorf("stack specs cannot be used with terraform commands; reference them from Environment or Service specs")
	default:
		return ctx, fmt.Errorf("unknown kind %q", kind)
	}

	ctx.env = env
	ctx.outDir, _ = filepath.Abs(filepath.Clean(ctx.outDir))
	return ctx, nil
}

func autoGenerateWorkspace(file, env, modulesRoot, out string, vars []string) (stackContext, string, error) {
	ctx, err := prepareWorkspaceContext(file, env, out)
	if err != nil {
		return ctx, "", err
	}
	cliVars, err := clihelper.ParseVarFlags(vars)
	if err != nil {
		return ctx, "", err
	}
	cliVars = clihelper.MergeVarMaps(clihelper.ParseVarEnv(), cliVars)
	embeddedRoot, customRoot, err := clihelper.ResolveModuleRoots(modulesRoot)
	if err != nil {
		return ctx, "", err
	}
	specPath := clihelper.DefaultString(file, "env.yaml")
	specDir := filepath.Dir(specPath)
	if ctx.kind == "Environment" {
		if err := generate.ExportEnvironmentWorkspaceForEnv(ctx.envCfg, embeddedRoot, customRoot, ctx.outDir, specDir, ctx.env, cliVars); err != nil {
			return ctx, "", err
		}
	} else {
		if err := generate.ExportServiceWorkspaceForEnv(ctx.svcCfg, ctx.envCfg, embeddedRoot, customRoot, ctx.outDir, specDir, ctx.env, cliVars); err != nil {
			return ctx, "", err
		}
	}
	tfvarsPath := filepath.Join(ctx.outDir, fmt.Sprintf("%s.tfvars", ctx.env))
	return ctx, tfvarsPath, nil
}

func runTfWithAction(action, file, env, modules, out string, vars []string, lockID string, opts clihelper.TerraformExecOptions) (retErr error) {
	ctx, tfvarsPath, err := autoGenerateWorkspace(file, env, modules, out, vars)
	if err != nil {
		return err
	}
	tfvarsArg := ""
	if tfvarsPath != "" {
		tfvarsArg = filepath.Base(tfvarsPath)
	}

	rootDir, _ := os.Getwd()

	_, finishRun := clihelper.StartLocalRun(action, file, ctx.env, ctx.outDir)
	defer func() {
		finishRun(retErr)
	}()

	if requiresEnvLock(action) {
		release, err := acquireEnvLock(ctx.envCfg.Metadata.Name, ctx.env)
		if err != nil {
			return err
		}
		defer release()
	}

	engine := defaultTfEngine
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	log := func(format string, args ...interface{}) {
		if stdout == nil {
			return
		}
		fmt.Fprintf(stdout, "[pltf] "+format+"\n", args...)
	}

	var session *daggerx.Session
	var imageCache *dagger.CacheVolume
	needImages := (action == "plan" || action == "apply") && hasImageBuilds(ctx)
	if needImages {
		session, err = daggerx.NewSession(clihelper.DaggerLogOutput(stderr))
		if err != nil {
			return err
		}
		defer session.Close()
		imageCache = session.Client.CacheVolume("pltf-image-cache")
		log("building images (push=%t)", action == "apply")
		if err := autoImageBuildWithSession(session, file, ctx.env, action == "apply", nil, imageCache); err != nil {
			return err
		}
	}
	tfRunner, err := terraform.NewRunner(ctx.outDir, stdout, stderr)
	if err != nil {
		return err
	}
	log("using workspace %s at %s", ctx.env, ctx.outDir)

	// Optional security scan happens before init/plan to fail fast.

	var scanSum *clihelper.TfsecSummary
	if action == "plan" && opts.Scan {
		var err error
		scanSum, err = clihelper.RunTfsecScan(ctx.outDir)
		if err != nil {
			return fmt.Errorf("tfsec scan failed: %w", err)
		}
	}

	envEntry := ctx.envCfg.Environments[ctx.env]
	bk, err := backend.Resolve(ctx.envCfg.Metadata.Provider, ctx.envCfg, envEntry)
	if err != nil {
		return err
	}
	if err := backend.Ensure(context.Background(), bk); err != nil {
		return fmt.Errorf("failed to ensure backend: %w", err)
	}

	log("running %s init", engine)
	if err := runTerraformInitWithRetry(tfRunner); err != nil {
		return fmt.Errorf("%s init failed: %w", engine, err)
	}
	log("ensuring workspace %s", ctx.env)
	if err := selectOrCreateWorkspace(tfRunner, ctx.env); err != nil {
		return err
	}

	common := func(args []string) []string {
		return clihelper.AppendTerraformArgs(args, opts)
	}

	var runErr error
	var planSum *clihelper.PlanSummary
	runStatus := "succeeded"
	var planResult planExecutionResult
	var planJSONPath string
	var costSum *clihelper.CostSummary
	if action == "plan" || action == "apply" || action == "destroy" {
		var errPlan error
		log("running %s plan", engine)
		planArgsFunc := func(args []string) []string {
			planOpts := opts
			planOpts.AutoApprove = false
			return clihelper.AppendTerraformArgs(args, planOpts)
		}
		planResult, errPlan = runTerraformPlan(tfRunner, ctx, tfvarsArg, opts, stderr, rootDir, planArgsFunc)
		if errPlan != nil {
			runErr = errPlan
		}
		if planResult.summary != nil {
			planSum = planResult.summary
		}
		if opts.DetailedExit && planResult.exitCode == 2 {
			runStatus = "changes"
		}
		if opts.Rover && planResult.summary != nil && planResult.summary.PlanJSON != "" {
			tfPath := engine
			if p, err := exec.LookPath(engine); err == nil {
				tfPath = p
			} else {
				fmt.Fprintf(stderr, "warn: %s not found in PATH for rover (defaulting to %q): %v\n", engine, tfPath, err)
			}
			r, err := rover.New(rover.Config{
				WorkingDir:   ctx.outDir,
				TfPath:       tfPath,
				PlanJSONPath: planResult.summary.PlanJSON,
				PlanPath:     planResult.planPathOnDisk,
			})
			if err != nil {
				fmt.Fprintf(stderr, "warn: rover init failed: %v\n", err)
			} else {
				if err := r.GenerateAssets(); err != nil {
					fmt.Fprintf(stderr, "warn: rover asset generation failed: %v\n", err)
				} else if err := r.StartServer("0.0.0.0:9000"); err != nil {
					fmt.Fprintf(stderr, "warn: rover server failed: %v\n", err)
				}
			}
		}
		if planResult.summary != nil {
			planJSONPath = planResult.summary.PlanJSON
		}
		if planResult.summary != nil && planJSONPath != "" && opts.Cost {
			if sum, err := runInfracost(planJSONPath, ctx.outDir); err == nil {
				costSum = sum
			} else {
				fmt.Fprintf(stderr, "warn: infracost run failed: %v\n", err)
			}
		}
	}
	if (action == "apply" || action == "destroy") && runErr != nil {
		return runErr
	}

	switch action {
	case "apply":
		log("running %s apply", engine)
		args := []string{"apply"}
		if tfvarsArg != "" {
			args = append(args, "-var-file="+tfvarsArg)
		}
		if _, _, err := tfRunner.Exec(common(args)); err != nil {
			runErr = fmtTfError(engine, "apply", err)
		}
	case "destroy":
		log("running %s destroy", engine)
		args := []string{"destroy"}
		if tfvarsArg != "" {
			args = append(args, "-var-file="+tfvarsArg)
		}
		if _, _, err := tfRunner.Exec(common(args)); err != nil {
			runErr = fmtTfError(engine, "destroy", err)
		}
	case "output":
		args := []string{"output"}
		if opts.JSONOutput {
			args = append(args, "-json")
		}
		if lockID != "" {
			args = append(args, lockID)
		}
		if _, _, err := tfRunner.Exec(args); err != nil {
			runErr = fmt.Errorf("%s output failed: %w", engine, err)
		}
	case "force-unlock":
		args := []string{"force-unlock", "-force", lockID}
		if _, _, err := tfRunner.Exec(args); err != nil {
			runErr = fmt.Errorf("%s force-unlock failed: %w", engine, err)
		}
	}

	if action == "plan" || action == "apply" {
		status := clihelper.RunSummary{
			Action: action,
			Spec:   file,
			Env:    env,
			OutDir: ctx.outDir,
			Plan:   planSum,
			Scan:   scanSum,
			Cost:   costSum,
		}
		if runErr != nil {
			status.Status = "failed"
			status.Err = runErr.Error()
		} else {
			status.Status = runStatus
		}
		if status.Plan != nil {
			status.AI = clihelper.MaybeAICritique(status)
		}
		if err := clihelper.MaybeUpsertPRComment(status); err != nil {
			fmt.Fprintf(stderr, "warn: failed to update PR comment: %v\n", err)
		}
	}

	return runErr
}

type planExecutionResult struct {
	summary        *clihelper.PlanSummary
	planArgs       []string
	exitCode       int
	planPathOnDisk string
}

func runTerraformPlan(runner *terraform.Runner, ctx stackContext, tfvarsArg string, opts clihelper.TerraformExecOptions, stderr io.Writer, rootDir string, appendArgs func([]string) []string) (planExecutionResult, error) {
	var res planExecutionResult
	args := []string{"plan"}
	if tfvarsArg != "" {
		args = append(args, "-var-file="+filepath.Join(ctx.outDir, tfvarsArg))
	}
	if opts.DetailedExit {
		args = append(args, "-detailed-exitcode")
	}
	planPath := opts.PlanFile
	planArg := opts.PlanFile
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
	planArgs := append([]string(nil), appendArgs(args)...)
	_, planExit, err := runner.Exec(planArgs)
	if err != nil && !(opts.DetailedExit && planExit == 2) {
		return res, fmt.Errorf("%s plan failed: %w", runner.EngineCmd(), err)
	}
	res.exitCode = planExit
	planPathOnDisk := planPath
	if !filepath.IsAbs(planPathOnDisk) {
		planPathOnDisk = filepath.Clean(planPathOnDisk)
		if !strings.HasPrefix(planPathOnDisk, ctx.outDir) {
			planPathOnDisk = filepath.Join(ctx.outDir, planPathOnDisk)
		}
	}
	res.planPathOnDisk = planPathOnDisk
	planJSONPath := ""
	var planJSONOutput string
	if tempPlan || strings.TrimSpace(planPathOnDisk) != "" {
		planJSONPath = strings.TrimSuffix(planPathOnDisk, filepath.Ext(planPathOnDisk)) + ".json"
		out, _, err := runner.Exec([]string{"show", "-json", planArg})
		if err != nil {
			fmt.Fprintf(stderr, "warn: %s show -json failed: %v\n", runner.EngineCmd(), err)
		} else {
			planJSONOutput = out
			if err := os.WriteFile(planJSONPath, []byte(out), 0o644); err != nil {
				fmt.Fprintf(stderr, "warn: write plan json failed: %v\n", err)
			}
		}
	}
	if sum, err := clihelper.CollectPlanSummaryWithRunner(ctx.outDir, planPathOnDisk, planJSONOutput, stderr); err == nil {
		res.summary = sum
		res.summary.RawPlanArgs = clihelper.SanitizePlanArgs(planArgs, ctx.outDir, tfvarsArg, rootDir)
		if planJSONPath != "" {
			res.summary.PlanJSON = planJSONPath
		}
	} else {
		fmt.Fprintf(stderr, "warn: failed to collect plan summary: %v\n", err)
	}
	res.planArgs = planArgs
	return res, nil
}

func runTerraformInitWithRetry(r *terraform.Runner) error {
	return clihelper.RunWithRetry(3, time.Second, func() error {
		_, _, err := r.Exec([]string{"init"})
		if err != nil && !clihelper.IsTransientInitError(err) {
			return err
		}
		return err
	})
}

func selectOrCreateWorkspace(r *terraform.Runner, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("workspace name is empty")
	}
	if _, _, err := r.Exec([]string{"workspace", "select", name}); err == nil {
		return nil
	}
	if _, _, err := r.Exec([]string{"workspace", "new", name}); err != nil {
		return fmt.Errorf("%s workspace create failed: %w", r.EngineCmd(), err)
	}
	return nil
}

func renderSnippet(root string, rng defsecTypes.Range) string {
	start := rng.GetStartLine()
	end := rng.GetEndLine()
	filename := rng.GetFilename()
	local := rng.GetLocalFilename()
	path := filepath.Join(root, filename)
	if _, err := os.Stat(path); err != nil {
		alt := filepath.Join(root, local)
		if _, err2 := os.Stat(alt); err2 == nil {
			path = alt
		} else {
			return ""
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	if start <= 0 {
		start = 1
	}
	if end <= 0 || end > len(lines) {
		end = len(lines)
	}
	var b strings.Builder
	for i := start; i <= end; i++ {
		fmt.Fprintf(&b, "      %4d | %s\n", i, lines[i-1])
	}
	return b.String()
}

func tfsecExitCode(metrics tfscanner.Metrics) int {
	if metrics.Executor.Counts.Failed == 0 {
		return 0
	}
	if metrics.Executor.Counts.Failed == metrics.Executor.Counts.Low {
		return 2
	}
	return 1
}

func printTfsecInsights(summary *clihelper.TfsecSummary) {
	report := summary.Report
	if strings.TrimSpace(report) == "" {
		report = clihelper.FormatTfsecReport(summary)
	}
	fmt.Fprint(os.Stderr, report)
}

func runInfracost(planJSONPath, workdir string) (*clihelper.CostSummary, error) {
	if _, err := exec.LookPath("infracost"); err != nil {
		return nil, fmt.Errorf("infracost binary not found in PATH")
	}
	if strings.TrimSpace(planJSONPath) == "" {
		return nil, fmt.Errorf("plan json path is empty")
	}
	args := []string{"breakdown", "--path", planJSONPath, "--format", "json"}
	out, err := clihelper.RunCmdOutput(workdir, "infracost", args...)
	if err != nil {
		return nil, err
	}
	sum := &clihelper.CostSummary{Raw: out}
	if t := extractInfracostTotal(out); t != "" {
		sum.TotalMonthly = t
	}
	if txt, err := clihelper.RunCmdOutput(workdir, "infracost", "breakdown", "--path", planJSONPath, "--format", "table"); err == nil {
		sum.Breakdown = txt
	}
	return sum, nil
}

func extractInfracostTotal(jsonStr string) string {
	type total struct {
		TotalMonthlyCost string `json:"totalMonthlyCost"`
	}
	type root struct {
		Projects []struct {
			Breakdown total `json:"breakdown"`
		} `json:"projects"`
		Summary total `json:"summary"`
	}
	var r root
	if err := json.Unmarshal([]byte(jsonStr), &r); err != nil {
		return ""
	}
	if r.Summary.TotalMonthlyCost != "" {
		return r.Summary.TotalMonthlyCost
	}
	for _, p := range r.Projects {
		if p.Breakdown.TotalMonthlyCost != "" {
			return p.Breakdown.TotalMonthlyCost
		}
	}
	return ""
}

func fmtTfError(engine, action string, err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "error asking for approval") {
		return fmt.Errorf("%s %s failed: Terraform requested interactive approval but stdin is unavailable; rerun with --auto-approve", engine, action)
	}
	return fmt.Errorf("%s %s failed: %w", engine, action, err)
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
	applyCmd.Flags().BoolVar(&applyAutoApprove, "auto-approve", true, "Pass -auto-approve to terraform apply")

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
	destroyCmd.Flags().BoolVar(&destroyAutoApprove, "auto-approve", true, "Pass -auto-approve to terraform destroy")

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
	planCmd.Flags().BoolVar(&planRover, "rover", false, "Run rover (https://github.com/yindia/rover) against the generated plan.json (requires rover binary in PATH)")
	planCmd.Flags().BoolVar(&planScan, "scan", false, "Run tfsec security scan against the generated Terraform")
	planCmd.Flags().BoolVar(&planCost, "cost", false, "Run infracost breakdown against the plan (requires infracost binary in PATH and INFRACOST_API_KEY)")

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
