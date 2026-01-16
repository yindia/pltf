package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"dagger.io/dagger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"pltf/pkg/config"
	"pltf/pkg/daggerx"
)

var (
	imageFile   string
	imageEnv    string
	imagePush   bool
	imageFilter []string
)

var curlyContentPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
var (
	imgSecretPrefix     = "PLTF_IMG_SECRET_"
	imgSecretFilePrefix = "PLTF_IMG_SECRET_FILE_"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Args:  cobra.NoArgs,
	Short: "Build and push Docker images defined in Environment or Service specs",
}

var imageBuildCmd = &cobra.Command{
	Use:   "build",
	Args:  cobra.NoArgs,
	Short: "Build Docker images from Environment or Service specs",
	Long: `Build Docker images defined in the spec with Dagger. Tags can include placeholders like
${env_name}, ${layer_name}, ${account_id}, ${region}.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		imageFile = defaultString(imageFile, "env.yaml")
		imageFile = cleanOptionalPath(imageFile)
		imageEnv = strings.TrimSpace(imageEnv)
		if err := ensureFile(imageFile, "spec file"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return autoImageBuild(imageFile, imageEnv, imagePush, imageFilter)
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageBuildCmd)

	imageBuildCmd.Flags().StringVarP(&imageFile, "file", "f", "env.yaml", "Path to the Environment or Service YAML file")
	imageBuildCmd.Flags().StringVarP(&imageEnv, "env", "e", "", "Environment key (dev, prod, etc.); required for both env and service specs")
	imageBuildCmd.Flags().BoolVar(&imagePush, "push", false, "Push images after build (overrides per-image push: true)")
	imageBuildCmd.Flags().StringArrayVar(&imageFilter, "image", nil, "Only build specific image name(s); can be repeated")
}

func autoImageBuild(file, env string, push bool, filter []string) error {
	session, err := daggerx.NewSession(daggerLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer session.Close()
	cache := session.Client.CacheVolume("pltf-image-cache")
	return autoImageBuildWithSession(session, file, env, push, filter, cache)
}

func autoImageBuildWithSession(session *daggerx.Session, file, env string, push bool, filter []string, cache *dagger.CacheVolume) error {
	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(abs)

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		return buildImages(session, envCfg.Images, baseDir, envName, envCfg.Environments[envName], nil, envCfg.Metadata.Name, push, filter, cache)
	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		return buildImages(session, svcCfg.Images, baseDir, envName, envCfg.Environments[envName], svcCfg, svcCfg.Metadata.Name, push, filter, cache)
	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment or Service)", file)
	}
}

func buildImages(session *daggerx.Session, images []config.ImageBuild, baseDir, envKey string, envEntry config.EnvironmentEntry, svcCfg *config.ServiceConfig, layerName string, push bool, filter []string, cache *dagger.CacheVolume) error {
	if len(images) == 0 {
		fmt.Fprintln(os.Stdout, "No images defined in spec.")
		return nil
	}

	filterSet := make(map[string]struct{})
	for _, name := range filter {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		filterSet[name] = struct{}{}
	}

	replacer := buildImageReplacer(envKey, envEntry, svcCfg, layerName)
	selected := make([]config.ImageBuild, 0, len(images))
	for _, img := range images {
		if len(filterSet) > 0 {
			if _, ok := filterSet[img.Name]; !ok {
				continue
			}
		}
		selected = append(selected, img)
	}
	if len(filterSet) > 0 && len(selected) == 0 {
		return fmt.Errorf("no matching images found for: %s", strings.Join(sortedKeys(filterSet), ", "))
	}
	if len(selected) == 0 {
		return fmt.Errorf("no images selected for build")
	}

	ctx := session.Ctx
	client := session.Client

	group, gctx := errgroup.WithContext(ctx)
	for _, img := range selected {
		img := img
		group.Go(func() error {
			return buildImageWithDagger(gctx, client, img, baseDir, replacer, push, cache)
		})
	}
	return group.Wait()
}

func buildImageWithDagger(ctx context.Context, client *dagger.Client, img config.ImageBuild, baseDir string, repl *strings.Replacer, push bool, cache *dagger.CacheVolume) error {
	contextPath, err := resolveImagePath(baseDir, replacePlaceholders(img.Context, repl))
	if err != nil {
		return fmt.Errorf("image %q context: %w", img.Name, err)
	}

	dockerfile := strings.TrimSpace(img.Dockerfile)
	if dockerfile != "" {
		dockerfile = replacePlaceholders(dockerfile, repl)
		if !filepath.IsAbs(dockerfile) {
			dockerfile = filepath.Join(contextPath, dockerfile)
		}
		dockerfile = filepath.Clean(dockerfile)
		rel, err := filepath.Rel(contextPath, dockerfile)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("image %q dockerfile must be within context %s", img.Name, contextPath)
		}
		dockerfile = rel
	}

	tags := make([]string, 0, len(img.Tags))
	for _, tag := range img.Tags {
		tag = replacePlaceholders(tag, repl)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	buildArgs := replaceMapValues(img.BuildArgs, repl)
	labels := replaceMapValues(img.Labels, repl)
	target := replacePlaceholders(img.Target, repl)
	include := replaceSliceValues(img.Include, repl)
	exclude := replaceSliceValues(img.Exclude, repl)

	opts := dagger.DirectoryDockerBuildOpts{}
	if dockerfile != "" {
		opts.Dockerfile = dockerfile
	}
	if target != "" {
		opts.Target = target
	}
	if len(buildArgs) > 0 {
		args := make([]dagger.BuildArg, 0, len(buildArgs))
		for key, val := range buildArgs {
			args = append(args, dagger.BuildArg{Name: key, Value: val})
		}
		opts.BuildArgs = args
	}
	if secrets, err := collectImageSecrets(client); err != nil {
		return err
	} else if len(secrets) > 0 {
		opts.Secrets = secrets
	}

	fmt.Fprintf(os.Stdout, "Building image %s (dagger)\n", img.Name)
	dir := client.Host().Directory(contextPath, dagger.HostDirectoryOpts{
		Include: include,
		Exclude: exclude,
	})
	platforms := resolveImagePlatforms(img)
	if len(platforms) == 0 {
		return fmt.Errorf("image %q has no platforms to build", img.Name)
	}

	containers := make([]*dagger.Container, len(platforms))
	group, _ := errgroup.WithContext(ctx)
	for idx, platform := range platforms {
		idx, platform := idx, platform
		group.Go(func() error {
			buildOpts := dagger.DirectoryDockerBuildOpts{
				Dockerfile: opts.Dockerfile,
				Target:     opts.Target,
				BuildArgs:  opts.BuildArgs,
				Secrets:    opts.Secrets,
			}
			if platform != "" {
				buildOpts.Platform = dagger.Platform(platform)
			}
			container := dir.DockerBuild(buildOpts)
			if cache != nil {
				container = container.WithMountedCache("/dagger/image-cache", cache)
			}
			for key, val := range labels {
				container = container.WithLabel(key, val)
			}
			containers[idx] = container
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("image %q failed to produce a build", img.Name)
	}

	primary := containers[0]
	var variants []*dagger.Container
	if len(containers) > 1 {
		variants = containers[1:]
	}

	for _, container := range containers {
		if container == nil {
			continue
		}
		if _, err := container.Sync(ctx); err != nil {
			return err
		}
	}

	if push {
		if len(tags) == 0 {
			return fmt.Errorf("image %q push requires at least one tag", img.Name)
		}
		for _, tag := range tags {
			fmt.Fprintf(os.Stdout, "Pushing %s\n", tag)
			publishOpts := []dagger.ContainerPublishOpts{}
			if len(variants) > 0 {
				publishOpts = append(publishOpts, dagger.ContainerPublishOpts{
					PlatformVariants: variants,
				})
			}
			if _, err := primary.Publish(ctx, tag, publishOpts...); err != nil {
				return err
			}
		}
		return nil
	}

	if len(tags) == 0 {
		return nil
	}

	for _, tag := range tags {
		fmt.Fprintf(os.Stdout, "Exporting %s to local image store\n", tag)
		if err := primary.ExportImage(ctx, tag); err != nil {
			return err
		}
	}
	return nil
}

func collectImageSecrets(client *dagger.Client) ([]*dagger.Secret, error) {
	var secrets []*dagger.Secret
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		if strings.HasPrefix(key, imgSecretPrefix) {
			name := strings.TrimPrefix(key, imgSecretPrefix)
			if name == "" || strings.TrimSpace(val) == "" {
				return nil, fmt.Errorf("%s%s is empty", imgSecretPrefix, name)
			}
			secrets = append(secrets, client.SetSecret(name, val))
			continue
		}
		if strings.HasPrefix(key, imgSecretFilePrefix) {
			name := strings.TrimPrefix(key, imgSecretFilePrefix)
			if name == "" || strings.TrimSpace(val) == "" {
				return nil, fmt.Errorf("%s%s is empty", imgSecretFilePrefix, name)
			}
			content, err := os.ReadFile(val)
			if err != nil {
				return nil, fmt.Errorf("%s%s read failed: %w", imgSecretFilePrefix, name, err)
			}
			secrets = append(secrets, client.SetSecret(name, string(content)))
		}
	}
	return secrets, nil
}

func buildImageReplacer(envKey string, envEntry config.EnvironmentEntry, svcCfg *config.ServiceConfig, layerName string) *strings.Replacer {
	if svcCfg != nil {
		layerName = svcCfg.Metadata.Name
	}
	placeholders := map[string]string{
		"env_name":    envKey,
		"layer_name":  layerName,
		"parent_name": layerName,
		"account_id":  envEntry.Account,
		"project_id":  envEntry.Account,
		"region":      envEntry.Region,
	}
	return strings.NewReplacer(buildPlaceholderPairs(placeholders)...)
}

func replaceMapValues(in map[string]string, repl *strings.Replacer) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, val := range in {
		out[key] = replacePlaceholders(val, repl)
	}
	return out
}

func replaceSliceValues(in []string, repl *strings.Replacer) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, val := range in {
		val = replacePlaceholders(val, repl)
		if val != "" {
			out = append(out, val)
		}
	}
	return out
}

func resolveImagePath(baseDir, p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("path is empty")
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}
	return filepath.Clean(p), nil
}

func replacePlaceholders(val string, repl *strings.Replacer) string {
	if val == "" {
		return ""
	}
	val = normalizeCurlyPlaceholders(val)
	return repl.Replace(val)
}

func buildPlaceholderPairs(vars map[string]string) []string {
	pairs := make([]string, 0, len(vars)*4)
	for key, value := range vars {
		pairs = append(pairs,
			"${"+key+"}", value,
			"${{"+key+"}}", value,
			"{"+key+"}", value,
			"{{"+key+"}}", value,
		)
	}
	return pairs
}

func normalizeCurlyPlaceholders(val string) string {
	var b strings.Builder
	for i := 0; i < len(val); {
		if val[i] == '{' && (i == 0 || val[i-1] != '$') {
			if i >= 2 && val[i-1] == '{' && val[i-2] == '$' {
				b.WriteByte(val[i])
				i++
				continue
			}
			if i+1 < len(val) && val[i+1] == '{' {
				endRel := strings.Index(val[i+2:], "}}")
				if endRel >= 0 {
					end := i + 2 + endRel
					content := val[i+2 : end]
					if curlyContentPattern.MatchString(content) {
						b.WriteString("${")
						b.WriteString(content)
						b.WriteString("}")
						i = end + 2
						continue
					}
				}
			}

			endRel := strings.IndexByte(val[i:], '}')
			if endRel > 1 {
				end := i + endRel
				content := val[i+1 : end]
				if curlyContentPattern.MatchString(content) {
					b.WriteString("${")
					b.WriteString(content)
					b.WriteString("}")
					i = end + 1
					continue
				}
			}
		}
		b.WriteByte(val[i])
		i++
	}
	return b.String()
}

func resolveImagePlatforms(img config.ImageBuild) []string {
	var platforms []string
	seen := make(map[string]struct{}, len(img.Platforms))
	for _, platform := range img.Platforms {
		platform = strings.TrimSpace(platform)
		if platform == "" {
			continue
		}
		if _, ok := seen[platform]; ok {
			continue
		}
		seen[platform] = struct{}{}
		platforms = append(platforms, platform)
	}
	if len(platforms) == 0 {
		platforms = append(platforms, fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	}
	return platforms
}
