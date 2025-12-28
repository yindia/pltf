package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"pltf/modules"
	"pltf/pkg/config"
)

func parseVarFlags(pairs []string) (map[string]string, error) {
	out := make(map[string]string)
	for _, p := range pairs {
		// allow key=value
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --var %q, expected key=value", p)
		}
		key := strings.TrimSpace(parts[0])
		value := parts[1] // keep as-is; parse later in generate
		if key == "" {
			return nil, fmt.Errorf("invalid --var %q, key cannot be empty", p)
		}
		out[key] = value
	}
	return out, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func cleanOptionalPath(path string) string {
	if path == "" {
		return path
	}
	return filepath.Clean(path)
}

func ensureFile(path, description string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s %q does not exist", description, path)
		}
		return fmt.Errorf("unable to read %s %q: %w", description, path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s %q is a directory, expected a file", description, path)
	}
	return nil
}

func ensureDir(path, description string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s %q does not exist", description, path)
		}
		return fmt.Errorf("unable to read %s %q: %w", description, path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s %q is not a directory", description, path)
	}
	return nil
}

func backupIfExists(path string, overwrite bool) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, expected a file", path)
	}
	if overwrite {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed removing %s: %w", path, err)
		}
		return nil
	}
	backup := fmt.Sprintf("%s.bak-%d", path, time.Now().Unix())
	if err := os.Rename(path, backup); err != nil {
		return fmt.Errorf("failed to backup %s to %s: %w", path, backup, err)
	}
	return nil
}

var (
	embeddedModulesOnce sync.Once
	embeddedModulesPath string
	embeddedModulesErr  error

	profileOnce sync.Once
	profileData *profileConfig
	profileErr  error
)

// resolveModulesRoot returns the modules root to use. If userPath is set, it is validated.
// Otherwise, embedded modules are materialized to a temp dir and used as default.
func resolveModulesRoot(userPath string) (string, error) {
	if strings.TrimSpace(userPath) != "" {
		userPath = filepath.Clean(userPath)
		if err := ensureDir(userPath, "modules root"); err != nil {
			return "", err
		}
		return userPath, nil
	}

	if prof := loadProfile(); prof != nil && strings.TrimSpace(prof.ModulesRoot) != "" {
		root := filepath.Clean(prof.ModulesRoot)
		if err := ensureDir(root, "modules root"); err == nil {
			return root, nil
		}
	}

	embeddedModulesOnce.Do(func() {
		embeddedModulesPath, embeddedModulesErr = modules.Materialize()
	})
	if embeddedModulesErr != nil {
		return "", embeddedModulesErr
	}
	return embeddedModulesPath, nil
}

// resolveModuleRoots returns embedded root plus optional custom root.
func resolveModuleRoots(userPath string) (embedded string, custom string, err error) {
	embedded, err = resolveModulesRoot("")
	if err != nil {
		return "", "", err
	}
	// Profile can set default modules_root; CLI flag wins.
	if strings.TrimSpace(userPath) == "" {
		if p := loadProfile(); p != nil && strings.TrimSpace(p.ModulesRoot) != "" {
			userPath = p.ModulesRoot
		}
	}
	if strings.TrimSpace(userPath) != "" {
		custom = filepath.Clean(userPath)
		if err := ensureDir(custom, "modules root"); err != nil {
			return "", "", err
		}
	}
	return embedded, custom, nil
}

func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCmdExit(dir, name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(interface{ ExitStatus() int }); ok {
			return status.ExitStatus(), err
		}
	}
	return 1, err
}

func runCmdOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func appendTfCommonArgs(args []string, opts tfExecOpts) []string {
	if opts.noColor {
		args = append(args, "-no-color")
	}
	if !opts.input {
		args = append(args, "-input=false")
	}
	if !opts.lock {
		args = append(args, "-lock=false")
	}
	if opts.lockTimeout != "" {
		args = append(args, "-lock-timeout="+opts.lockTimeout)
	}
	if opts.parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", opts.parallelism))
	}
	for _, t := range opts.targets {
		args = append(args, "-target="+t)
	}
	if opts.refresh != nil {
		args = append(args, fmt.Sprintf("-refresh=%t", *opts.refresh))
	}
	return args
}

type backendDetails struct {
	typeName      string
	bucket        string
	container     string
	resourceGroup string
	region        string
}

func isS3Backend(b backendDetails) bool {
	switch b.typeName {
	case "", "aws", "s3":
		return true
	default:
		return false
	}
}

func computeBackend(envCfg *config.EnvironmentConfig, envName string) (backendDetails, error) {
	envEntry, ok := envCfg.Environments[envName]
	if !ok {
		return backendDetails{}, fmt.Errorf("environment %q not found", envName)
	}
	bType := strings.ToLower(strings.TrimSpace(envCfg.Backend.Type))
	if bType == "" {
		bType = strings.ToLower(envCfg.Metadata.Provider)
	}

	bucket := envCfg.Backend.Bucket
	container := envCfg.Backend.Container
	resourceGroup := envCfg.Backend.ResourceGroup
	region := envCfg.Backend.Region
	if region == "" {
		region = envEntry.Region
	}

	switch bType {
	case "aws", "s3", "":
		if bucket == "" {
			if envEntry.Region == "" {
				return backendDetails{}, fmt.Errorf("environment %q region is required for backend naming", envName)
			}
			bucket = fmt.Sprintf("%s-%s-%s", envCfg.Metadata.Name, envCfg.Metadata.Org, envEntry.Region)
		}
	case "gcp", "google", "gcs":
		if bucket == "" {
			bucket = fmt.Sprintf("%s-%s-%s", envCfg.Metadata.Name, envCfg.Metadata.Org, envEntry.Region)
		}
	case "azure", "azurerm":
		if bucket == "" {
			return backendDetails{}, fmt.Errorf("backend.bucket (storage account name) is required for azure backend")
		}
		if container == "" {
			container = "tfstate"
		}
		if resourceGroup == "" {
			resourceGroup = fmt.Sprintf("%s-tfstate-rg", envCfg.Metadata.Name)
		}
	default:
		return backendDetails{}, fmt.Errorf("unsupported backend type %q", bType)
	}

	return backendDetails{
		typeName:      bType,
		bucket:        bucket,
		container:     container,
		resourceGroup: resourceGroup,
		region:        region,
	}, nil
}

func ensureS3Bucket(bucket, region string) error {
	if bucket == "" {
		return fmt.Errorf("backend bucket is empty")
	}

	ctx := context.Background()
	cfg, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if region != "" {
			o.Region = region
		}
	})

	// Head bucket
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &bucket})
	if err == nil {
		return nil
	}

	createInput := &s3.CreateBucketInput{Bucket: &bucket}
	if region != "" && region != "us-east-1" {
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	if _, err := client.CreateBucket(ctx, createInput); err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
	}
	return nil
}

// selectEnvName chooses an environment name from input, env var, or config context.
// For Environment specs: if env is empty, fall back to PLTF_DEFAULT_ENV, then to the only environment defined.
// For Service specs: env must exist in both envCfg and svcCfg; fallbacks mirror above.
func selectEnvName(kind string, env string, envCfg *config.EnvironmentConfig, svcCfg *config.ServiceConfig) (string, error) {
	candidate := strings.TrimSpace(env)
	if candidate == "" {
		if prof := loadProfile(); prof != nil && strings.TrimSpace(prof.DefaultEnv) != "" {
			candidate = strings.TrimSpace(prof.DefaultEnv)
		}
	}
	if candidate == "" {
		if def := strings.TrimSpace(os.Getenv("PLTF_DEFAULT_ENV")); def != "" {
			candidate = def
		}
	}

	switch kind {
	case "Environment":
		if candidate != "" {
			if _, ok := envCfg.Environments[candidate]; ok {
				return candidate, nil
			}
			return "", fmt.Errorf("environment %q not found in spec; available: %s", candidate, strings.Join(sortedKeys(envCfg.Environments), ","))
		}
		if len(envCfg.Environments) == 1 {
			for k := range envCfg.Environments {
				return k, nil
			}
		}
		return "", fmt.Errorf("--env is required (set flag or PLTF_DEFAULT_ENV)")
	case "Service":
		if candidate != "" {
			if _, ok := envCfg.Environments[candidate]; !ok {
				return "", fmt.Errorf("environment %q not found in Environment; available: %s", candidate, strings.Join(sortedKeys(envCfg.Environments), ","))
			}
			if _, ok := svcCfg.Metadata.EnvRef[candidate]; !ok {
				return "", fmt.Errorf("environment %q not found in service envRef; available: %s", candidate, strings.Join(sortedKeys(svcCfg.Metadata.EnvRef), ","))
			}
			return candidate, nil
		}
		if def := strings.TrimSpace(os.Getenv("PLTF_DEFAULT_ENV")); def != "" {
			if _, ok := envCfg.Environments[def]; ok {
				if _, ok := svcCfg.Metadata.EnvRef[def]; ok {
					return def, nil
				}
			}
		}
		if len(svcCfg.Metadata.EnvRef) == 1 {
			for k := range svcCfg.Metadata.EnvRef {
				return k, nil
			}
		}
		return "", fmt.Errorf("--env is required (set flag or PLTF_DEFAULT_ENV)")
	default:
		return "", fmt.Errorf("unsupported kind %q for env selection", kind)
	}
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type profileConfig struct {
	ModulesRoot string `yaml:"modules_root"`
	DefaultEnv  string `yaml:"default_env"`
	DefaultOut  string `yaml:"default_out"`
	Telemetry   bool   `yaml:"telemetry"`
}


func loadProfile() *profileConfig {
	profileOnce.Do(func() {
		path := os.Getenv("PLTF_PROFILE")
		if strings.TrimSpace(path) == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				profileErr = err
				return
			}
			path = filepath.Join(home, ".pltf", "profile.yaml")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				profileErr = err // File exists but is unreadable
			}
			return
		}
		var cfg profileConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			profileErr = fmt.Errorf("failed to parse profile %s: %w", path, err)
			return
		}
		profileData = &cfg
	})
	if profileErr != nil {
		// A corrupt profile is a non-fatal warning, not a hard error.
		fmt.Fprintf(os.Stderr, "warn: unable to load profile: %v\n", profileErr)
		return nil
	}
	return profileData
}
