package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"pltf/modules"
	"pltf/pkg/config"
	"pltf/pkg/runner"
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

func parseVarEnv() map[string]string {
	out := make(map[string]string)
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		if strings.HasPrefix(key, "PLTF_VAR_") {
			name := strings.TrimPrefix(key, "PLTF_VAR_")
			if strings.TrimSpace(name) == "" {
				continue
			}
			out[name] = value
		}
	}
	return out
}

func mergeVarMaps(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
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
		return resolveModulesRootFromRef(userPath)
	}

	if prof := loadProfile(); prof != nil && strings.TrimSpace(prof.ModulesRoot) != "" {
		root, err := resolveModulesRootFromRef(prof.ModulesRoot)
		if err == nil {
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

// resolveModulesRootWithLabel returns the modules root and a human-friendly label.
func resolveModulesRootWithLabel(userPath string) (string, string, error) {
	if strings.TrimSpace(userPath) != "" {
		root, err := resolveModulesRootFromRef(userPath)
		return root, userPath, err
	}

	if prof := loadProfile(); prof != nil && strings.TrimSpace(prof.ModulesRoot) != "" {
		root, err := resolveModulesRootFromRef(prof.ModulesRoot)
		if err == nil {
			return root, prof.ModulesRoot, nil
		}
	}

	root, err := resolveModulesRoot("")
	if err != nil {
		return "", "", err
	}
	return root, "embedded modules", nil
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
		custom, err = resolveModulesRootFromRef(userPath)
		if err != nil {
			return "", "", err
		}
	}
	return embedded, custom, nil
}

func resolveModulesRootFromRef(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", nil
	}
	resolved := ref
	if strings.Contains(ref, "://") {
		baseDir, _ := os.Getwd()
		path, err := config.ResolveGitRef(ref, baseDir)
		if err != nil {
			return "", err
		}
		resolved = path
	}
	resolved = filepath.Clean(resolved)
	if err := ensureDir(resolved, "modules root"); err != nil {
		return "", err
	}
	return resolved, nil
}

func selectModuleMeta(module config.Module, embeddedMetas, customMetas map[string]*config.ModuleMetadata, embeddedRoot string) (*config.ModuleMetadata, error) {
	if strings.EqualFold(module.Source, "custom") {
		if len(customMetas) == 0 {
			return nil, fmt.Errorf("module %q (type=%s) marked source=custom but no custom modules root provided", module.ID, module.Type)
		}
		meta, ok := customMetas[module.Type]
		if !ok {
			return nil, fmt.Errorf("module %q (type=%s) marked source=custom but metadata not found", module.ID, module.Type)
		}
		return meta, nil
	}

	meta, ok := embeddedMetas[module.Type]
	if !ok {
		if len(customMetas) > 0 {
			if alt, ok := customMetas[module.Type]; ok {
				return alt, nil
			}
		}
		return nil, fmt.Errorf("module %q type %q not found in embedded modules (%s); use source: custom with --modules", module.ID, module.Type, embeddedRoot)
	}
	return meta, nil
}

func runCmdOutputWithIO(dir, name string, stderr io.Writer, args ...string) (string, error) {
	return runner.Default.RunOutput(runner.Cmd{
		Name:   name,
		Args:   args,
		Dir:    dir,
		Stderr: stderr,
	})
}

func runWithRetry(attempts int, baseDelay time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	delay := baseDelay
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < attempts-1 {
				time.Sleep(delay)
				delay *= 2
			}
			continue
		}
		return nil
	}
	return lastErr
}

func isTransientInitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	transient := []string{
		"timeout",
		"timed out",
		"connection reset",
		"connection refused",
		"temporary",
		"temporarily unavailable",
		"i/o timeout",
		"eof",
		"tls handshake timeout",
		"network",
		"dial tcp",
		"lookup",
	}
	for _, s := range transient {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func runCmdOutput(dir, name string, args ...string) (string, error) {
	return runCmdOutputWithIO(dir, name, os.Stderr, args...)
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
	if opts.autoApprove {
		args = append(args, "-auto-approve")
	}
	return args
}

func daggerLogOutput(stderr io.Writer) io.Writer {
	if stderr == nil {
		return os.Stderr
	}
	return stderr
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
