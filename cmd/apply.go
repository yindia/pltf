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

	"github.com/aquasecurity/defsec/pkg/extrafs"
	tfscanner "github.com/aquasecurity/defsec/pkg/scanners/terraform"
	defsecTypes "github.com/aquasecurity/defsec/pkg/types"
	"github.com/spf13/cobra"

	"pltf/pkg/backend"
	"pltf/pkg/config"
	"pltf/pkg/daggerx"
	"pltf/pkg/generate"
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
  pltf terraform plan -f service.yaml -e dev --detailed-exitcode --plan-file=/tmp/plan.tfplan
  pltf terraform plan -f env.yaml -e prod --rover   # renders plan.json and opens rover (https://github.com/yindia/rover)
  pltf terraform plan -f env.yaml -e prod --scan    # run tfsec against generated TF
  pltf terraform plan -f env.yaml -e prod --cost    # run infracost breakdown (if infracost binary present)`,
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
			rover:        planRover,
			scan:         planScan,
			cost:         planCost,
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
	rover        bool
	scan         bool
	cost         bool
	stdout       io.Writer
	stderr       io.Writer
}

type stackContext struct {
	kind   string
	env    string
	envCfg *config.EnvironmentConfig
	svcCfg *config.ServiceConfig
	outDir string
}

func prepareWorkspaceContext(file, env, out string) (stackContext, error) {
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
		ctx.svcCfg = nil
		if out == "" {
			ctx.outDir = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
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
	cliVars, err := parseVarFlags(vars)
	if err != nil {
		return ctx, "", err
	}
	cliVars = mergeVarMaps(parseVarEnv(), cliVars)
	embeddedRoot, customRoot, err := resolveModuleRoots(modulesRoot)
	if err != nil {
		return ctx, "", err
	}
	specPath := defaultString(file, "env.yaml")
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

func runTfWithAction(action, file, env, modules, out string, vars []string, lockID string, opts tfExecOpts) (retErr error) {
	ctx, tfvarsPath, err := autoGenerateWorkspace(file, env, modules, out, vars)
	if err != nil {
		return err
	}
	tfvarsArg := ""
	if tfvarsPath != "" {
		tfvarsArg = filepath.Join("/work", filepath.Base(tfvarsPath))
	}

	_, finishRun := startLocalRun(action, file, ctx.env, ctx.outDir)
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
	stdout := opts.stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	session, err := daggerx.NewSession(daggerLogOutput(stderr))
	if err != nil {
		return err
	}
	defer session.Close()

	imageCache := session.Client.CacheVolume("pltf-image-cache")
	tfPluginCache := session.Client.CacheVolume("pltf-terraform-plugin-cache")
	tfRunner := newTfDaggerRunner(session, ctx.outDir, stdout, stderr, tfPluginCache)

	log := func(format string, args ...interface{}) {
		if stdout == nil {
			return
		}
		fmt.Fprintf(stdout, "[pltf] "+format+"\n", args...)
	}
	log("using workspace %s at %s", ctx.env, ctx.outDir)

	// Optional security scan happens before init/plan to fail fast.

	var scanSum *tfsecSummary
	if action == "plan" && opts.scan {
		var err error
		scanSum, err = runTfsecScan(ctx.outDir)
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
	if err := runTerraformInitWithRetryDagger(tfRunner); err != nil {
		return fmt.Errorf("%s init failed: %w", engine, err)
	}
	log("ensuring workspace %s", ctx.env)
	if err := selectOrCreateWorkspaceDagger(tfRunner, ctx.env); err != nil {
		return err
	}

	common := func(args []string) []string {
		args = appendTfCommonArgs(args, opts)
		return args
	}

	var runErr error
	var planSum *planSummary
	runStatus := "succeeded"
	var planResult planExecutionResult
	var planJSONPath string
	var costSum *costSummary
	if action == "plan" || action == "apply" || action == "destroy" {
		log("building images (push=%t)", action == "apply")
		pushImages := action == "apply"
		if err := autoImageBuildWithSession(session, file, ctx.env, pushImages, nil, imageCache); err != nil {
			return err
		}
		var errPlan error
		log("running %s plan", engine)
		planResult, errPlan = runTerraformPlan(tfRunner, ctx, tfvarsArg, opts, stderr, common)
		if errPlan != nil {
			runErr = errPlan
		}
		if planResult.summary != nil {
			planSum = planResult.summary
		}
		if opts.detailedExit && planResult.exitCode == 2 {
			runStatus = "changes"
		}
		if opts.rover && planResult.summary != nil && planResult.summary.PlanJSON != "" {
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
		if planResult.summary != nil && planJSONPath != "" && opts.cost {
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
		args := []string{tfRunner.engineCmd(), "apply"}
		if tfvarsArg != "" {
			args = append(args, "-var-file="+tfvarsArg)
		}
		if _, _, err := tfRunner.exec(common(args), true); err != nil {
			runErr = fmtTfError(engine, "apply", err)
		}
	case "destroy":
		log("running %s destroy", engine)
		args := []string{tfRunner.engineCmd(), "destroy"}
		if tfvarsArg != "" {
			args = append(args, "-var-file="+tfvarsArg)
		}
		if _, _, err := tfRunner.exec(common(args), true); err != nil {
			runErr = fmtTfError(engine, "destroy", err)
		}
	case "output":
		args := []string{tfRunner.engineCmd(), "output"}
		if opts.jsonOutput {
			args = append(args, "-json")
		}
		if lockID != "" {
			args = append(args, lockID)
		}
		if _, _, err := tfRunner.exec(args, false); err != nil {
			runErr = fmt.Errorf("%s output failed: %w", engine, err)
		}
	case "force-unlock":
		args := []string{tfRunner.engineCmd(), "force-unlock", "-force", lockID}
		if _, _, err := tfRunner.exec(args, false); err != nil {
			runErr = fmt.Errorf("%s force-unlock failed: %w", engine, err)
		}
	}

	if action == "plan" || action == "apply" {
		status := tfRunSummary{
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
			status.AI = maybeAICritique(status)
		}
		if err := maybeUpsertPRComment(status); err != nil {
			fmt.Fprintf(stderr, "warn: failed to update PR comment: %v\n", err)
		}
	}

	return runErr
}

type tfsecFinding struct {
	Severity    string
	Rule        string
	Location    string
	Description string
	Impact      string
	Resolution  string
	Links       []string
	Snippet     string
}

type tfsecSummary struct {
	ExitCode int
	Failed   int
	Low      int
	Medium   int
	High     int
	Critical int
	Findings []tfsecFinding
	Report   string
	Timings  struct {
		DiskIO     time.Duration
		Parsing    time.Duration
		Adaptation time.Duration
		Checks     time.Duration
		Total      time.Duration
	}
	Counts struct {
		ModulesDownloaded int
		ModulesProcessed  int
		BlocksProcessed   int
		FilesRead         int
		Passed            int
		Ignored           int
	}
}

type costSummary struct {
	TotalMonthly string
	Breakdown    string
	Raw          string
}

func runTfsecScan(dir string) (*tfsecSummary, error) {
	root := filepath.Clean(dir)
	scnr := tfscanner.New()
	results, metrics, err := scnr.ScanFSWithMetrics(context.Background(), extrafs.OSDir(root), ".")
	if err != nil {
		return nil, err
	}
	exit := tfsecExitCode(metrics)
	summary := &tfsecSummary{
		ExitCode: exit,
		Failed:   metrics.Executor.Counts.Failed,
		Low:      metrics.Executor.Counts.Low,
		Medium:   metrics.Executor.Counts.Medium,
		High:     metrics.Executor.Counts.High,
		Critical: metrics.Executor.Counts.Critical,
	}
	summary.Counts.ModulesDownloaded = metrics.Parser.Counts.ModuleDownloads
	summary.Counts.ModulesProcessed = metrics.Parser.Counts.Modules
	summary.Counts.BlocksProcessed = metrics.Parser.Counts.Blocks
	summary.Counts.FilesRead = metrics.Parser.Counts.Files
	summary.Counts.Passed = metrics.Executor.Counts.Passed
	summary.Counts.Ignored = metrics.Executor.Counts.Ignored
	summary.Timings.DiskIO = metrics.Parser.Timings.DiskIODuration
	summary.Timings.Parsing = metrics.Parser.Timings.ParseDuration
	summary.Timings.Adaptation = metrics.Executor.Timings.Adaptation
	summary.Timings.Checks = metrics.Executor.Timings.RunningChecks
	summary.Timings.Total = metrics.Timings.Total
	for _, res := range results {
		status := strings.ToLower(fmt.Sprint(res.Status()))
		if !strings.Contains(status, "fail") {
			continue
		}
		f := tfsecFinding{
			Severity:    string(res.Severity()),
			Rule:        res.Rule().LongID(),
			Location:    res.Range().String(),
			Description: res.Description(),
			Impact:      res.Rule().Impact,
			Resolution:  res.Rule().Resolution,
			Links:       res.Rule().Links,
			Snippet:     renderSnippet(root, res.Range()),
		}
		summary.Findings = append(summary.Findings, f)
	}
	summary.Report = formatTfsecReport(summary)

	if exit != 0 {
		fmt.Fprintf(os.Stderr, "warn: tfsec reported issues (exit=%d, failed=%d low=%d medium=%d high=%d critical=%d)\n",
			exit, summary.Failed, summary.Low, summary.Medium, summary.High, summary.Critical)
	}
	fmt.Fprint(os.Stderr, summary.Report)
	return summary, nil
}

type planExecutionResult struct {
	summary        *planSummary
	planArgs       []string
	exitCode       int
	planPathOnDisk string
}

func runTerraformPlan(runner *tfDaggerRunner, ctx stackContext, tfvarsArg string, opts tfExecOpts, stderr io.Writer, common func([]string) []string) (planExecutionResult, error) {
	var res planExecutionResult
	args := []string{runner.engineCmd(), "plan"}
	if tfvarsArg != "" {
		args = append(args, "-var-file="+tfvarsArg)
	}
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
	planArgs := append([]string(nil), common(args)...)
	_, planExit, err := runner.exec(planArgs, true)
	if err != nil && !(opts.detailedExit && planExit == 2) {
		return res, fmt.Errorf("%s plan failed: %w", runner.engineCmd(), err)
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
	if tempPlan || strings.TrimSpace(planPathOnDisk) != "" {
		planJSONPath = strings.TrimSuffix(planPathOnDisk, filepath.Ext(planPathOnDisk)) + ".json"
		out, _, err := runner.exec([]string{runner.engineCmd(), "show", "-json", planArg}, false)
		if err != nil {
			fmt.Fprintf(stderr, "warn: %s show -json failed: %v\n", runner.engineCmd(), err)
		} else if err := os.WriteFile(planJSONPath, []byte(out), 0o644); err != nil {
			fmt.Fprintf(stderr, "warn: write plan json failed: %v\n", err)
		}
	}
	if sum, err := collectPlanSummaryWithRunner(runner, ctx.outDir, planArg, planPathOnDisk, stderr); err == nil {
		res.summary = sum
		res.summary.RawPlanArgs = planArgs
		if planJSONPath != "" {
			res.summary.PlanJSON = planJSONPath
		}
	} else {
		fmt.Fprintf(stderr, "warn: failed to collect plan summary: %v\n", err)
	}
	res.planArgs = planArgs
	return res, nil
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

func printTfsecInsights(summary *tfsecSummary) {
	fmt.Fprint(os.Stderr, formatTfsecInsights(summary))
}

func formatDurationMs(d time.Duration) string {
	ms := float64(d) / float64(time.Millisecond)
	return fmt.Sprintf("%.6fms", ms)
}

func formatTfsecInsights(summary *tfsecSummary) string {
	var b strings.Builder
	b.WriteString("  timings\n")
	b.WriteString("  ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  disk i/o             %s\n", formatDurationMs(summary.Timings.DiskIO))
	fmt.Fprintf(&b, "  parsing              %s\n", formatDurationMs(summary.Timings.Parsing))
	fmt.Fprintf(&b, "  adaptation           %s\n", formatDurationMs(summary.Timings.Adaptation))
	fmt.Fprintf(&b, "  checks               %s\n", formatDurationMs(summary.Timings.Checks))
	fmt.Fprintf(&b, "  total                %s\n\n", formatDurationMs(summary.Timings.Total))

	b.WriteString("  counts\n")
	b.WriteString("  ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  modules downloaded   %d\n", summary.Counts.ModulesDownloaded)
	fmt.Fprintf(&b, "  modules processed    %d\n", summary.Counts.ModulesProcessed)
	fmt.Fprintf(&b, "  blocks processed     %d\n", summary.Counts.BlocksProcessed)
	fmt.Fprintf(&b, "  files read           %d\n\n", summary.Counts.FilesRead)

	b.WriteString("  results\n")
	b.WriteString("  ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  passed               %d\n", summary.Counts.Passed)
	fmt.Fprintf(&b, "  ignored              %d\n", summary.Counts.Ignored)
	fmt.Fprintf(&b, "  critical             %d\n", summary.Critical)
	fmt.Fprintf(&b, "  high                 %d\n", summary.High)
	fmt.Fprintf(&b, "  medium               %d\n", summary.Medium)
	fmt.Fprintf(&b, "  low                  %d\n\n", summary.Low)

	totalProblems := summary.Failed
	fmt.Fprintf(&b, "  %d passed, %d ignored, %d potential problem(s) detected.\n", summary.Counts.Passed, summary.Counts.Ignored, totalProblems)
	return b.String()
}

func formatTfsecReport(summary *tfsecSummary) string {
	var b strings.Builder
	for i, f := range summary.Findings {
		fmt.Fprintf(&b, "Result #%d %s %s\n", i+1, strings.ToUpper(f.Severity), f.Description)
		b.WriteString("────────────────────────────────────────────────────────────────────────────────\n")
		if f.Location != "" {
			fmt.Fprintf(&b, "  %s\n", f.Location)
			b.WriteString("────────────────────────────────────────────────────────────────────────────────\n")
		}
		if f.Snippet != "" {
			b.WriteString(f.Snippet)
			if !strings.HasSuffix(f.Snippet, "\n") {
				b.WriteString("\n")
			}
		}
		if f.Rule != "" {
			fmt.Fprintf(&b, "          ID %s\n", f.Rule)
		}
		if f.Impact != "" {
			fmt.Fprintf(&b, "      Impact %s\n", f.Impact)
		}
		if f.Resolution != "" {
			fmt.Fprintf(&b, "  Resolution %s\n", f.Resolution)
		}
		if len(f.Links) > 0 {
			b.WriteString("\n  More Information\n")
			for _, link := range f.Links {
				fmt.Fprintf(&b, "  - %s\n", link)
			}
		}
		b.WriteString("────────────────────────────────────────────────────────────────────────────────\n\n")
	}
	b.WriteString(formatTfsecInsights(summary))
	return b.String()
}

func runInfracost(planJSONPath, workdir string) (*costSummary, error) {
	if _, err := exec.LookPath("infracost"); err != nil {
		return nil, fmt.Errorf("infracost binary not found in PATH")
	}
	if strings.TrimSpace(planJSONPath) == "" {
		return nil, fmt.Errorf("plan json path is empty")
	}
	args := []string{"breakdown", "--path", planJSONPath, "--format", "json"}
	out, err := runCmdOutput(workdir, "infracost", args...)
	if err != nil {
		return nil, err
	}
	sum := &costSummary{Raw: out}
	if t := extractInfracostTotal(out); t != "" {
		sum.TotalMonthly = t
	}
	if txt, err := runCmdOutput(workdir, "infracost", "breakdown", "--path", planJSONPath, "--format", "table"); err == nil {
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
