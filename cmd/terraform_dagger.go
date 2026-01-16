package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dagger.io/dagger"

	"pltf/pkg/daggerx"
)

type tfDaggerRunner struct {
	ctx         context.Context
	client      *dagger.Client
	workdir     string
	stdout      io.Writer
	stderr      io.Writer
	container   *dagger.Container
	pluginCache *dagger.CacheVolume
}

const terraformRcContent = `
plugin_cache_dir = "/work/.terraform-plugin-cache"
`

func newTfDaggerRunner(session *daggerx.Session, workdir string, stdout, stderr io.Writer, pluginCache *dagger.CacheVolume) *tfDaggerRunner {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &tfDaggerRunner{
		ctx:         session.Ctx,
		client:      session.Client,
		workdir:     workdir,
		stdout:      stdout,
		stderr:      stderr,
		pluginCache: pluginCache,
	}
}

func (r *tfDaggerRunner) engineCmd() string {
	return "terraform"
}

func (r *tfDaggerRunner) imageRef() string {
	return "ghcr.io/yindia/terraform-cli:latest"
}

func (r *tfDaggerRunner) baseContainer() (*dagger.Container, error) {
	ctr := r.client.Container().From(r.imageRef()).
		WithDirectory("/work", r.client.Host().Directory(r.workdir)).
		WithWorkdir("/work").
		WithUser("root").
		WithExec([]string{"chown", "-R", "nonroot", "/work"})

	envs, files, secrets, err := collectTerraformEnv()
	if err != nil {
		return nil, err
	}
	for key, val := range envs {
		ctr = ctr.WithEnvVariable(key, val)
	}
	if len(files) > 0 {
		ctr, err = mountTerraformFiles(r.client, ctr, envs, files)
		if err != nil {
			return nil, err
		}
	}
	if len(secrets.envSecrets) > 0 || len(secrets.fileSecrets) > 0 {
		ctr, err = applyTerraformSecrets(r.client, ctr, secrets)
		if err != nil {
			return nil, err
		}
	}

	if r.pluginCache != nil {
		ctr = ctr.WithMountedCache("/work/.terraform-plugin-cache", r.pluginCache)
		ctr = ctr.WithExec([]string{"mkdir", "-p", "/work/.terraform-plugin-cache"})
		ctr = ctr.WithExec([]string{"chown", "-R", "nonroot", "/work/.terraform-plugin-cache"})
		ctr = ctr.WithEnvVariable("TF_PLUGIN_CACHE_DIR", "/work/.terraform-plugin-cache")
		ctr = ctr.WithNewFile("/home/nonroot/.terraformrc", terraformRcContent, dagger.ContainerWithNewFileOpts{
			Permissions: 0o644,
			Owner:       "nonroot:nonroot",
		})
	}

	home, err := os.UserHomeDir()
	if err == nil {
		// mount host credentials into the nonroot home so Terraform can read them
		if info, e := os.Stat(filepath.Join(home, ".aws")); e == nil && info.IsDir() {
			ctr = withHostDirIfExists(r.client, ctr, "/home/nonroot/.aws", filepath.Join(home, ".aws"))
			ctr = ctr.WithExec([]string{"chown", "-R", "nonroot", "/home/nonroot/.aws"})
		}
		if info, e := os.Stat(filepath.Join(home, ".docker")); e == nil && info.IsDir() {
			ctr = withHostDirIfExists(r.client, ctr, "/home/nonroot/.docker", filepath.Join(home, ".docker"))
			ctr = ctr.WithExec([]string{"chown", "-R", "nonroot", "/home/nonroot/.docker"})
		}
		if info, e := os.Stat(filepath.Join(home, ".config", "gcloud")); e == nil && info.IsDir() {
			ctr = withHostDirIfExists(r.client, ctr, "/pltf/.config/gcloud", filepath.Join(home, ".config", "gcloud"))
		}
		if info, e := os.Stat(filepath.Join(home, ".azure")); e == nil && info.IsDir() {
			ctr = withHostDirIfExists(r.client, ctr, "/pltf/.azure", filepath.Join(home, ".azure"))
		}
	}

	ctr = withAwsAuth(ctr)

	ctr = ctr.WithEnvVariable("HOME", "/home/nonroot")

	return ctr.WithUser("nonroot"), nil
}

func withHostDirIfExists(client *dagger.Client, ctr *dagger.Container, target, source string) *dagger.Container {
	info, err := os.Stat(source)
	if err != nil || !info.IsDir() {
		return ctr
	}
	return ctr.WithDirectory(target, client.Host().Directory(source))
}

func withAwsAuth(ctr *dagger.Container) *dagger.Container {
	awsEnvVars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"AWS_SECURITY_TOKEN",
		"AWS_REGION",
		"AWS_DEFAULT_REGION",
		"AWS_PROFILE",
	}

	for _, envVar := range awsEnvVars {
		if value, ok := os.LookupEnv(envVar); ok {
			ctr = ctr.WithEnvVariable(envVar, value)
		}
	}
	return ctr
}

const (
	tfFilePrefix       = "PLTF_TF_FILE_"
	tfSecretPrefix     = "PLTF_TF_SECRET_"
	tfSecretFilePrefix = "PLTF_TF_SECRET_FILE_"
)

type tfSecrets struct {
	envSecrets  map[string]string
	fileSecrets map[string]string
}

func collectTerraformEnv() (map[string]string, map[string]string, tfSecrets, error) {
	envs := map[string]string{}
	files := map[string]string{}
	secrets := tfSecrets{
		envSecrets:  map[string]string{},
		fileSecrets: map[string]string{},
	}
	// do not forward host temp/runtime envs that can point to host filesystem
	skip := map[string]bool{"TMPDIR": true, "TMP": true, "TEMP": true, "XDG_RUNTIME_DIR": true}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		if skip[key] {
			continue
		}
		if key != "" {
			envs[key] = val
		}
		if strings.HasPrefix(key, tfFilePrefix) {
			name := strings.TrimPrefix(key, tfFilePrefix)
			if name != "" {
				files[name] = val
			}
			continue
		}
		if strings.HasPrefix(key, tfSecretPrefix) {
			name := strings.TrimPrefix(key, tfSecretPrefix)
			if name != "" {
				secrets.envSecrets[name] = val
			}
			continue
		}
		if strings.HasPrefix(key, tfSecretFilePrefix) {
			name := strings.TrimPrefix(key, tfSecretFilePrefix)
			if name != "" {
				secrets.fileSecrets[name] = val
			}
		}
	}
	for name, path := range files {
		if strings.TrimSpace(path) == "" {
			return nil, nil, tfSecrets{}, fmt.Errorf("%s%s is empty", tfFilePrefix, name)
		}
		if _, err := os.Stat(path); err != nil {
			return nil, nil, tfSecrets{}, fmt.Errorf("%s%s path not found: %w", tfFilePrefix, name, err)
		}
	}
	for name, path := range secrets.fileSecrets {
		if strings.TrimSpace(path) == "" {
			return nil, nil, tfSecrets{}, fmt.Errorf("%s%s is empty", tfSecretFilePrefix, name)
		}
		if _, err := os.Stat(path); err != nil {
			return nil, nil, tfSecrets{}, fmt.Errorf("%s%s path not found: %w", tfSecretFilePrefix, name, err)
		}
	}
	return envs, files, secrets, nil
}

func mountTerraformFiles(client *dagger.Client, ctr *dagger.Container, envs map[string]string, files map[string]string) (*dagger.Container, error) {
	for name, path := range files {
		target := filepath.Join("/pltf/secrets", name)
		abs := path
		if !filepath.IsAbs(abs) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			abs = filepath.Join(cwd, path)
		}
		abs = filepath.Clean(abs)
		ctr = ctr.WithMountedFile(target, client.Host().File(abs))
		ctr = ctr.WithEnvVariable(name, target)
	}
	return ctr, nil
}

func applyTerraformSecrets(client *dagger.Client, ctr *dagger.Container, secrets tfSecrets) (*dagger.Container, error) {
	for name, val := range secrets.envSecrets {
		if strings.TrimSpace(val) == "" {
			return nil, fmt.Errorf("%s%s is empty", tfSecretPrefix, name)
		}
		secret := client.SetSecret(name, val)
		ctr = ctr.WithSecretVariable(name, secret)
	}
	for name, path := range secrets.fileSecrets {
		abs := path
		if !filepath.IsAbs(abs) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			abs = filepath.Join(cwd, path)
		}
		abs = filepath.Clean(abs)
		content, err := os.ReadFile(abs)
		if err != nil {
			return nil, fmt.Errorf("%s%s read failed: %w", tfSecretFilePrefix, name, err)
		}
		secret := client.SetSecret(name, string(content))
		target := filepath.Join("/run/secrets", name)
		ctr = ctr.WithMountedSecret(target, secret)
		ctr = ctr.WithEnvVariable(name, target)
	}
	return ctr, nil
}

func (r *tfDaggerRunner) ensureContainer() (*dagger.Container, error) {
	if r.container != nil {
		return r.container, nil
	}
	ctr, err := r.baseContainer()
	if err != nil {
		return nil, err
	}
	r.container = ctr
	return ctr, nil
}

func (r *tfDaggerRunner) exec(args []string, export bool) (string, int, error) {
	base, err := r.ensureContainer()
	if err != nil {
		return "", 1, err
	}
	ctr := base.WithExec(args)

	out, err := ctr.Stdout(r.ctx)
	exit := 0
	if err != nil {
		var execErr *dagger.ExecError
		if errors.As(err, &execErr) {
			exit = execErr.ExitCode
			if execErr.Stdout != "" && r.stdout != nil {
				_, _ = r.stdout.Write([]byte(execErr.Stdout))
			}
			if execErr.Stderr != "" && r.stderr != nil {
				_, _ = r.stderr.Write([]byte(execErr.Stderr))
			}
		} else {
			return "", exit, err
		}
	} else if out != "" && r.stdout != nil {
		_, _ = r.stdout.Write([]byte(out))
	}

	if err == nil {
		if stderrOut, stderrErr := ctr.Stderr(r.ctx); stderrErr == nil && stderrOut != "" && r.stderr != nil {
			_, _ = r.stderr.Write([]byte(stderrOut))
		} else if stderrErr != nil {
			return out, 1, stderrErr
		}
	}

	if export {
		if _, err := ctr.Directory("/work").Export(r.ctx, r.workdir); err != nil {
			return out, exit, err
		}
	}

	r.container = ctr

	return out, exit, err
}

func runTerraformInitWithRetryDagger(r *tfDaggerRunner) error {
	return runWithRetry(3, time.Second, func() error {
		_, _, err := r.exec([]string{r.engineCmd(), "init"}, true)
		if err != nil && !isTransientInitError(err) {
			return err
		}
		return err
	})
}

func selectOrCreateWorkspaceDagger(r *tfDaggerRunner, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("workspace name is empty")
	}
	if _, _, err := r.exec([]string{r.engineCmd(), "workspace", "select", name}, true); err == nil {
		return nil
	}
	if _, _, err := r.exec([]string{r.engineCmd(), "workspace", "new", name}, true); err != nil {
		return fmt.Errorf("%s workspace create failed: %w", r.engineCmd(), err)
	}
	return nil
}
