package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"pltf/pkg/augment"
	"pltf/pkg/config"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

var (
	moduleRefPattern     = regexp.MustCompile(`^module\.([a-zA-Z0-9_.-]+)\.([a-zA-Z0-9_]+)$`)
	varRefPattern        = regexp.MustCompile(`^var\.([a-zA-Z0-9_]+)$`)
	parentRefPattern     = regexp.MustCompile(`^parent\.([a-zA-Z0-9_]+)$`)
	interpolationPattern = regexp.MustCompile(`\$\{([^}]+)\}`)
	templatePattern      = regexp.MustCompile(`\$\{\{([^}]+)\}\}`)
	fullExprPattern      = regexp.MustCompile(`^\s*\$\{\{?\s*(.+?)\s*\}?\}\s*$`)
	moduleRefAnywhere    = regexp.MustCompile(`module\.([a-zA-Z0-9_.-]+)\.[a-zA-Z0-9_]+`)
	curlyContentPattern  = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	identifierPattern    = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

type moduleScope string

const (
	scopeEnv     moduleScope = "env"
	scopeService moduleScope = "service"
)

type Generator struct {
	envCfg              *config.EnvironmentConfig
	svcCfg              *config.ServiceConfig // can be nil
	embeddedModulesRoot string
	customModulesRoot   string
	envName             string // metadata.name of the Environment
	envKey              string // selected environment entry key (e.g., dev/prod)
	outDir              string
	specDir             string // directory of the spec file (env/service), used for copying file inputs
	cliVars             map[string]string

	isService bool
	envEntry  config.EnvironmentEntry

	// service only
	svcEnvEntry config.ServiceEnvRefEntry

	// all modules in scope: env + service for service stacks
	allModules []config.Module

	// cache of metadata for all modules in the stack
	modMap map[string]*config.ModuleMetadata
	// map module type -> source root
	moduleRootByType map[string]string

	// map output name -> list of module IDs that provide it
	outputProviders map[string][]string

	// map module id -> metadata
	moduleMetas map[string]*config.ModuleMetadata

	// map module id -> scope (env or service)
	moduleScopes map[string]moduleScope

	// merged locals/vars for this run (env + service + CLI)
	mergedVars map[string]interface{}

	iamAugmentations map[string]augment.Augmentation

	globalLabels map[string]string

	// module dependencies (module id -> set of module ids it depends on)
	moduleDeps map[string]map[string]struct{}
	// cache where git module repos are stored
	moduleGitCacheDir string

	// workspace mode: render locals/providers with var references
	useWorkspace bool
}

// NewGenerator builds a Generator with loaded module metadata, output wiring, and env/service context.
// envCfg is required; svcCfg is optional (nil for environment stacks). envName must exist in envCfg.Environments
// and, if svcCfg is provided, in svcCfg.Metadata.EnvRef. modulesRoot should contain subdirs per module type with module.yaml.
func NewGenerator(
	envCfg *config.EnvironmentConfig,
	svcCfg *config.ServiceConfig, // can be nil
	embeddedRoot string,
	customRoot string,
	envName string,
	outDir string,
	specDir string,
	cliVars map[string]string,
) (*Generator, error) {
	// Normalize paths for cross-platform support (macOS/Linux/Windows).
	outDir = filepath.Clean(outDir)
	specDir = strings.TrimSpace(specDir)
	if specDir != "" {
		specDir = filepath.Clean(specDir)
	}
	envName = strings.TrimSpace(envName)
	if envName == "" {
		return nil, fmt.Errorf("environment name cannot be empty")
	}

	cacheDir := filepath.Join(".pltf", "cache", "modules")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create module git cache dir %s: %w", cacheDir, err)
	}

	g := &Generator{
		envCfg:              envCfg,
		svcCfg:              svcCfg,
		embeddedModulesRoot: filepath.Clean(embeddedRoot),
		envKey:              envName,
		outDir:              outDir,
		specDir:             specDir,
		cliVars:             cliVars,
		isService:           svcCfg != nil,
		moduleScopes:        map[string]moduleScope{},
		moduleDeps:          map[string]map[string]struct{}{},
	}
	g.moduleGitCacheDir = cacheDir

	if strings.TrimSpace(customRoot) != "" {
		g.customModulesRoot = filepath.Clean(customRoot)
	}

	// Get env
	envEntry, ok := envCfg.Environments[envName]
	if !ok {
		return nil, fmt.Errorf("environment %q not found in %s", envName, envCfg.Metadata.Name)
	}
	g.envEntry = envEntry
	g.envName = envCfg.Metadata.Name
	if strings.TrimSpace(g.envEntry.Region) == "" {
		return nil, fmt.Errorf("environment %q region is required", envName)
	}

	// Get service env if applicable
	if g.isService {
		svcEnvEntry, ok := svcCfg.Metadata.EnvRef[envName]
		if !ok {
			return nil, fmt.Errorf("service envRef.%q not found in service %q", envName, svcCfg.Metadata.Name)
		}
		g.svcEnvEntry = svcEnvEntry
		g.allModules = append(g.allModules, envCfg.Modules...)
		g.allModules = append(g.allModules, svcCfg.Modules...)

		for _, mod := range envCfg.Modules {
			g.moduleScopes[mod.ID] = scopeEnv
		}
		for _, mod := range svcCfg.Modules {
			g.moduleScopes[mod.ID] = scopeService
		}
	} else {
		g.allModules = envCfg.Modules
		for _, mod := range envCfg.Modules {
			g.moduleScopes[mod.ID] = scopeEnv
		}
	}

	// Load all module metadata
	moduleTypes := map[string]struct{}{}
	customTypes := map[string]struct{}{}
	g.modMap = make(map[string]*config.ModuleMetadata)
	g.moduleRootByType = make(map[string]string)

	for i := range g.allModules {
		mod := &g.allModules[i]
		source := strings.TrimSpace(mod.Source)
		if config.IsGitSource(source) {
			meta, moduleDir, err := config.LoadModuleMetadataFromGitSource(source, g.moduleGitCacheDir)
			if err != nil {
				return nil, fmt.Errorf("module %q git source %q: %w", mod.ID, source, err)
			}
			mod.Type = meta.Type
			moduleTypes[meta.Type] = struct{}{}
			g.modMap[meta.Type] = meta
			g.moduleRootByType[meta.Type] = moduleDir
			continue
		}
		if mod.Type == "" {
			return nil, fmt.Errorf("module %q must declare a type when source is not a git URL", mod.ID)
		}
		moduleTypes[mod.Type] = struct{}{}
		if config.IsCustomRootSource(source) {
			customTypes[mod.Type] = struct{}{}
		}
	}

	if len(customTypes) > 0 && g.customModulesRoot == "" {
		return nil, fmt.Errorf("modules marked source=custom require --modules or profile.modules_root")
	}

	typeList := make([]string, 0, len(moduleTypes))
	for t := range moduleTypes {
		typeList = append(typeList, t)
	}

	var roots []string
	// Priority: custom root first if present, then embedded
	if g.customModulesRoot != "" {
		roots = append(roots, g.customModulesRoot)
	}
	roots = append(roots, g.embeddedModulesRoot)

	modRecords, err := config.ScanModuleRoots(roots, typeList)
	if err != nil {
		return nil, fmt.Errorf("failed to scan modules roots %v: %w", roots, err)
	}
	for t, rec := range modRecords {
		// Skip types already loaded from git sources
		if _, ok := g.modMap[t]; ok {
			continue
		}
		// Enforce custom modules to come from custom root
		if _, ok := customTypes[t]; ok && g.customModulesRoot != "" && filepath.Clean(rec.Root) != filepath.Clean(g.customModulesRoot) {
			return nil, fmt.Errorf("module type %q marked source=custom but not found in custom root %s", t, g.customModulesRoot)
		}
		g.modMap[t] = rec.Meta
		g.moduleRootByType[t] = rec.Root
	}

	// Pre-process to find all output providers
	g.outputProviders = make(map[string][]string)
	g.moduleMetas = make(map[string]*config.ModuleMetadata)

	for _, m := range g.allModules {
		meta, ok := g.modMap[m.Type]
		if !ok {
			return nil, fmt.Errorf("module type %q (id=%s) not found in module roots", m.Type, m.ID)
		}
		g.moduleMetas[m.ID] = meta

		for _, out := range meta.Outputs {
			g.outputProviders[out.Name] = append(g.outputProviders[out.Name], m.ID)
		}
	}

	g.addGlobalTags()
	g.mergedVars = g.getMergedVars()
	serviceName := envName
	if svcCfg != nil {
		serviceName = svcCfg.Metadata.Name
	}
	g.iamAugmentations = augment.Build(augment.Context{
		Provider:    envCfg.Metadata.Provider,
		EnvName:     envName,
		ServiceName: serviceName,
		IsService:   g.isService,
		Modules:     g.allModules,
		Vars:        g.mergedVars,
		ModuleMetas: g.moduleMetas,
		ModuleScopes: func() map[string]string {
			out := make(map[string]string, len(g.moduleScopes))
			for id, scope := range g.moduleScopes {
				out[id] = string(scope)
			}
			return out
		}(),
	})

	return g, nil
}

// EnableWorkspaceMode switches the generator to workspace export behavior.
func (g *Generator) EnableWorkspaceMode() {
	g.useWorkspace = true
}

// Generate writes Terraform for the configured stack into g.outDir.
func (g *Generator) Generate() (err error) {
	if err := g.assertSafeOutDir(); err != nil {
		return err
	}
	var restoreTerraform func() error
	if restoreTerraform, err = preserveTerraformCache(g.outDir); err != nil {
		return err
	}
	if restoreTerraform != nil {
		defer func() {
			if restoreErr := restoreTerraform(); restoreErr != nil && err == nil {
				err = restoreErr
			}
		}()
	}
	// 1. Create output directory
	if err := os.RemoveAll(g.outDir); err != nil {
		return fmt.Errorf("failed to clear output dir %s: %w", g.outDir, err)
	}
	if err := os.MkdirAll(g.outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output dir %s: %w", g.outDir, err)
	}

	// 2. Write shared files (versions.tf, providers.tf, secrets.tf)
	if err := g.writeBaseFiles(false, ""); err != nil {
		return err
	}

	// 3. Write a file for each module
	modulesToGen := g.envCfg.Modules
	if g.isService {
		modulesToGen = g.svcCfg.Modules
	}

	usedModuleTypes := make(map[string]bool)
	for _, m := range modulesToGen {
		meta := g.moduleMetas[m.ID]
		usedModuleTypes[meta.Type] = true

		mWithFiles, err := g.materializeFileInputs(g.applyAugmentations(m))
		if err != nil {
			return err
		}
		if err := g.writeModuleFile(mWithFiles, meta); err != nil {
			return err
		}
	}

	// 4. Copy module sources
	if err := copyUsedModules(g.outDir, usedModuleTypes, g.moduleRootByType); err != nil {
		return fmt.Errorf("failed to copy modules: %w", err)
	}

	// 5. Emit outputs.tf for all module outputs
	if err := g.writeOutputsFile(modulesToGen); err != nil {
		return fmt.Errorf("failed to write outputs.tf: %w", err)
	}

	return nil
}

// GenerateWorkspace writes Terraform using var-based locals for workspace exports.
func (g *Generator) GenerateWorkspace(workspaceKeyPrefix string) (err error) {
	if err := g.assertSafeOutDir(); err != nil {
		return err
	}
	var restoreTerraform func() error
	if restoreTerraform, err = preserveTerraformCache(g.outDir); err != nil {
		return err
	}
	if restoreTerraform != nil {
		defer func() {
			if restoreErr := restoreTerraform(); restoreErr != nil && err == nil {
				err = restoreErr
			}
		}()
	}
	// 1. Create output directory
	if err := os.RemoveAll(g.outDir); err != nil {
		return fmt.Errorf("failed to clear output dir %s: %w", g.outDir, err)
	}
	if err := os.MkdirAll(g.outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output dir %s: %w", g.outDir, err)
	}

	g.EnableWorkspaceMode()

	// 2. Write shared files (versions.tf, providers.tf, secrets.tf)
	if err := g.writeBaseFiles(true, workspaceKeyPrefix); err != nil {
		return err
	}

	// 3. Write a file for each module
	modulesToGen := g.envCfg.Modules
	if g.isService {
		modulesToGen = g.svcCfg.Modules
	}

	usedModuleTypes := make(map[string]bool)
	for _, m := range modulesToGen {
		meta := g.moduleMetas[m.ID]
		usedModuleTypes[meta.Type] = true

		mWithFiles, err := g.materializeFileInputs(g.applyAugmentations(m))
		if err != nil {
			return err
		}
		if err := g.writeModuleFile(mWithFiles, meta); err != nil {
			return err
		}
	}

	// 4. Copy module sources
	if err := copyUsedModules(g.outDir, usedModuleTypes, g.moduleRootByType); err != nil {
		return fmt.Errorf("failed to copy modules: %w", err)
	}

	// 5. Emit outputs.tf for all module outputs
	if err := g.writeOutputsFile(modulesToGen); err != nil {
		return fmt.Errorf("failed to write outputs.tf: %w", err)
	}

	return nil
}

// assertSafeOutDir prevents accidental deletion of ".", "/" or empty paths when cleaning the output dir.
func (g *Generator) assertSafeOutDir() error {
	if strings.TrimSpace(g.outDir) == "" {
		return fmt.Errorf("output directory is empty")
	}
	cleaned := filepath.Clean(g.outDir)
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return fmt.Errorf("refusing to operate on unsafe output directory %q", cleaned)
	}
	return nil
}

func (g *Generator) writeModuleFile(m config.Module, meta *config.ModuleMetadata) error {
	modFile := hclwrite.NewEmptyFile()
	body := modFile.Body()

	modBlock := body.AppendNewBlock("module", []string{m.ID})
	modBody := modBlock.Body()

	modBody.SetAttributeValue("source", cty.StringVal(fmt.Sprintf("./modules/%s", meta.Type)))

	// Process declared inputs
	for _, inSpec := range meta.Inputs {
		if err := g.processInput(modBody, m, inSpec); err != nil {
			return err
		}
	}

	// Process any extra inputs not in metadata
	for key, val := range m.Inputs {
		if !inputDeclared(meta, key) {
			g.collectDepsFromValue(m.ID, val)
			if err := g.setAttribute(modBody, key, val, true); err != nil {
				return fmt.Errorf("module %q extra input %q: %w", m.ID, key, err)
			}
		}
	}

	if deps := g.sortedDeps(m.ID); len(deps) > 0 {
		if g.isService {
			serviceDeps, _ := g.splitByScope(deps)
			deps = serviceDeps
		}
		if len(deps) > 0 {
			modBody.SetAttributeRaw("depends_on", g.dependsTokens(deps))
		}
	}

	modPath := filepath.Join(g.outDir, fmt.Sprintf("%s.tf", m.ID))
	return os.WriteFile(modPath, modFile.Bytes(), 0o644)
}

func (g *Generator) materializeFileInputs(m config.Module) (config.Module, error) {
	if m.Inputs == nil {
		return m, nil
	}

	updated := make(map[string]interface{}, len(m.Inputs))
	for k, v := range m.Inputs {
		nv, err := g.copyFileValues(v)
		if err != nil {
			return m, fmt.Errorf("module %q input %q: %w", m.ID, k, err)
		}
		updated[k] = nv
	}
	m.Inputs = updated
	return m, nil
}

func (g *Generator) copyFileValues(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return g.copyFileIfPath(val)
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			nv, err := g.copyFileValues(item)
			if err != nil {
				return nil, err
			}
			out[i] = nv
		}
		return out, nil
	case []string:
		out := make([]string, len(val))
		for i, item := range val {
			nv, err := g.copyFileIfPath(item)
			if err != nil {
				return nil, err
			}
			out[i] = nv
		}
		return out, nil
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, item := range val {
			nv, err := g.copyFileValues(item)
			if err != nil {
				return nil, err
			}
			out[k] = nv
		}
		return out, nil
	case map[string]string:
		out := make(map[string]string, len(val))
		for k, item := range val {
			nv, err := g.copyFileIfPath(item)
			if err != nil {
				return nil, err
			}
			out[k] = nv
		}
		return out, nil
	default:
		return v, nil
	}
}

func (g *Generator) copyFileIfPath(val string) (string, error) {
	s := strings.TrimSpace(val)
	if s == "" {
		return val, nil
	}
	if strings.Contains(s, "${") || strings.Contains(s, "{{") || strings.Contains(s, "://") || strings.HasPrefix(s, "module.") || strings.HasPrefix(s, "parent.") || strings.HasPrefix(s, "var.") {
		return val, nil
	}

	var src string
	if filepath.IsAbs(s) {
		src = s
	} else {
		if g.specDir == "" {
			return val, nil
		}
		src = filepath.Clean(filepath.Join(g.specDir, s))
	}

	info, err := os.Stat(src)
	if err != nil || info.IsDir() {
		return val, nil
	}

	rel := s
	if filepath.IsAbs(src) && g.specDir != "" {
		if r, err := filepath.Rel(g.specDir, src); err == nil && !strings.HasPrefix(r, "..") {
			rel = r
		} else {
			rel = filepath.Base(src)
		}
	}

	rel = filepath.Clean(rel)
	if strings.HasPrefix(rel, "..") {
		rel = filepath.Base(src)
	}

	dest := filepath.Join(g.outDir, rel)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return val, fmt.Errorf("create dir for %s: %w", dest, err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return val, fmt.Errorf("read file %s: %w", src, err)
	}
	if err := os.WriteFile(dest, data, info.Mode().Perm()); err != nil {
		return val, fmt.Errorf("copy file to %s: %w", dest, err)
	}

	return filepath.ToSlash(rel), nil
}

func (g *Generator) writeOutputsFile(mods []config.Module) error {
	outFile := hclwrite.NewEmptyFile()
	body := outFile.Body()

	seen := map[string]struct{}{}
	baseCounts := map[string]int{}

	for _, m := range mods {
		meta := g.moduleMetas[m.ID]
		if meta == nil {
			continue
		}
		for _, out := range meta.Outputs {
			base := outputLabelBase(out.Name)
			baseCounts[base]++
		}
	}

	for _, m := range mods {
		meta := g.moduleMetas[m.ID]
		if meta == nil {
			continue
		}
		for _, out := range meta.Outputs {
			name := g.uniqueOutputName(m.ID, out.Name, baseCounts, seen)
			seen[name] = struct{}{}

			block := body.AppendNewBlock("output", []string{name})
			b := block.Body()
			b.SetAttributeRaw("value", hclwrite.TokensForTraversal(moduleOutputTraversal(m.ID, out.Name)))
			if desc := strings.TrimSpace(out.Description); desc != "" {
				b.SetAttributeValue("description", cty.StringVal(desc))
			}
			if strings.EqualFold(out.Capability, "secret") {
				b.SetAttributeValue("sensitive", cty.BoolVal(true))
			}
		}
	}

	path := filepath.Join(g.outDir, "outputs.tf")
	return os.WriteFile(path, outFile.Bytes(), 0o644)
}

func (g *Generator) processInput(modBody *hclwrite.Body, m config.Module, inSpec config.InputSpec) error {
	val, err := g.resolveInput(modBody, m, inSpec)
	if err != nil {
		return err
	}

	g.collectDepsFromValue(m.ID, val)
	return g.setAttribute(modBody, inSpec.Name, val, true)
}

func (g *Generator) resolveInput(modBody *hclwrite.Body, m config.Module, inSpec config.InputSpec) (interface{}, error) {
	// 1. Direct input from YAML
	if raw, ok := m.Inputs[inSpec.Name]; ok {
		return raw, nil
	}

	// 2. Auto-fill platform fields
	switch inSpec.Name {
	case "env_name":
		return g.envKey, nil
	case "layer_name":
		layer := g.envName
		if g.isService {
			layer = g.svcCfg.Metadata.Name
		}
		return layer, nil
	case "module_name":
		return m.ID, nil
	}

	// 3. Auto-wire from another module's output
	if providers, ok := g.outputProviders[inSpec.Name]; ok {
		// Filter out the current module from the list of providers
		var candidates []string
		for _, p := range providers {
			if p != m.ID {
				candidates = append(candidates, p)
			}
		}

		if g.isService {
			serviceProviders, envProviders := g.splitByScope(candidates)
			if len(serviceProviders) > 1 {
				return nil, fmt.Errorf(
					"module %q input %q can be satisfied by multiple service modules: %v. Please specify which to use in your YAML.",
					m.ID, inSpec.Name, serviceProviders,
				)
			}
			if len(envProviders) > 1 && len(serviceProviders) == 0 {
				return nil, fmt.Errorf(
					"module %q input %q can be satisfied by multiple environment modules: %v. Please specify which to use in your YAML.",
					m.ID, inSpec.Name, envProviders,
				)
			}

			switch {
			case len(serviceProviders) == 1:
				g.addDep(m.ID, serviceProviders[0])
				setAttrModuleOutputRef(modBody, inSpec.Name, serviceProviders[0], inSpec.Name)
				return nil, nil // Attribute is set directly, so we return nil
			case len(envProviders) == 1:
				setAttrParentOutputRef(modBody, inSpec.Name, inSpec.Name)
				return nil, nil // Attribute is set directly, so we return nil
			}
		} else {
			if len(candidates) > 1 {
				return nil, fmt.Errorf(
					"module %q input %q can be satisfied by multiple modules: %v. Please specify which to use in your YAML.",
					m.ID, inSpec.Name, candidates,
				)
			}
			if len(candidates) == 1 {
				g.addDep(m.ID, candidates[0])
				setAttrModuleOutputRef(modBody, inSpec.Name, candidates[0], inSpec.Name)
				return nil, nil // Attribute is set directly, so we return nil
			}
		}
	}

	// 4. Wire from merged locals/vars if available
	if _, ok := g.mergedVars[inSpec.Name]; ok {
		return nil, g.setVarReference(modBody, inSpec.Name, inSpec.Name)
	}

	// 5. Handle required/default logic
	if inSpec.Required && inSpec.Default == nil {
		return nil, fmt.Errorf("module %q (type=%s) missing required input %q", m.ID, m.Type, inSpec.Name)
	}
	if inSpec.Default == nil {
		return nil, nil // Skip if no value and no default
	}
	return inSpec.Default, nil
}

func (g *Generator) setAttribute(body *hclwrite.Body, name string, value interface{}, replaceEnvLayer bool) error {
	if value == nil {
		return nil
	}
	value = g.replaceIntrinsicPlaceholdersInValue(value, replaceEnvLayer)
	tokens, err := g.valueToTokens(value)
	if err != nil {
		return fmt.Errorf("could not convert value for %q to HCL tokens: %w", name, err)
	}
	body.SetAttributeRaw(name, tokens)
	return nil
}

func (g *Generator) valueToTokens(value interface{}) (hclwrite.Tokens, error) {
	switch v := value.(type) {
	case string:
		return g.stringToTokens(v), nil
	case fmt.Stringer:
		return g.stringToTokens(v.String()), nil
	case []string:
		items := make([]interface{}, 0, len(v))
		for _, s := range v {
			items = append(items, s)
		}
		return g.sliceToTokens(items)
	case []map[string]interface{}:
		items := make([]interface{}, 0, len(v))
		for _, m := range v {
			items = append(items, m)
		}
		return g.sliceToTokens(items)
	case map[string]interface{}:
		return g.mapToTokens(v)
	case []interface{}:
		return g.sliceToTokens(v)
	default:
		// Fallback to cty for simple literals
		ctyVal, err := toCtyValue(v)
		if err != nil {
			return nil, err
		}
		return hclwrite.TokensForValue(ctyVal), nil
	}
}

func (g *Generator) replaceIntrinsicPlaceholdersInValue(value interface{}, replaceEnvLayer bool) interface{} {
	switch v := value.(type) {
	case string:
		return g.replaceIntrinsicPlaceholders(v, replaceEnvLayer)
	case fmt.Stringer:
		return g.replaceIntrinsicPlaceholders(v.String(), replaceEnvLayer)
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			out[i] = g.replaceIntrinsicPlaceholders(item, replaceEnvLayer)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = g.replaceIntrinsicPlaceholdersInValue(item, replaceEnvLayer)
		}
		return out
	case []map[string]interface{}:
		out := make([]map[string]interface{}, len(v))
		for i, item := range v {
			if converted, ok := g.replaceIntrinsicPlaceholdersInValue(item, replaceEnvLayer).(map[string]interface{}); ok {
				out[i] = converted
			}
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for key, val := range v {
			out[key] = g.replaceIntrinsicPlaceholdersInValue(val, replaceEnvLayer)
		}
		return out
	case map[string]string:
		out := make(map[string]string, len(v))
		for key, val := range v {
			out[key] = g.replaceIntrinsicPlaceholders(val, replaceEnvLayer)
		}
		return out
	default:
		return value
	}
}

func normalizeCurlyPlaceholders(val string) string {
	var b strings.Builder
	for i := 0; i < len(val); {
		// Only rewrite `{...}` when it's not already `${...}`
		if val[i] == '{' && (i == 0 || val[i-1] != '$') {
			// If this is the second `{` in a `${{...}}` sequence, leave it to the template replacer.
			if i >= 2 && val[i-1] == '{' && val[i-2] == '$' {
				b.WriteByte(val[i])
				i++
				continue
			}
			// Handle double-curly {{foo}}
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

			// Handle single-curly {foo}
			endRel := strings.IndexByte(val[i:], '}')
			if endRel > 1 { // must contain at least one char between {}
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

func tokensForObjectKey(key string) hclwrite.Tokens {
	if identifierPattern.MatchString(key) {
		return hclwrite.TokensForIdentifier(key)
	}
	return hclwrite.TokensForValue(cty.StringVal(key))
}

func outputLabelBase(outName string) string {
	base := strings.TrimSpace(outName)
	if base == "" {
		base = "output"
	}
	return base
}

func (g *Generator) uniqueOutputName(modID, outName string, baseCounts map[string]int, seen map[string]struct{}) string {
	base := outputLabelBase(outName)
	if baseCounts[base] > 1 {
		base = fmt.Sprintf("%s_%s", modID, base)
	}

	name := base
	i := 2
	for {
		if _, exists := seen[name]; !exists {
			return name
		}
		name = fmt.Sprintf("%s_%d", base, i)
		i++
	}
}

func (g *Generator) stringToTokens(s string) hclwrite.Tokens {
	normalized := templatePattern.ReplaceAllString(s, `${$1}`)

	// If the whole string is a single interpolation like ${...} or ${{...}}, render as pure expression (no quotes).
	if sm := fullExprPattern.FindStringSubmatch(normalized); sm != nil {
		return g.expressionToTokens(strings.TrimSpace(sm[1]))
	}

	// Case 1: The entire string is a single reference, e.g., "module.foo.bar"
	if sm := moduleRefPattern.FindStringSubmatch(normalized); sm != nil {
		return g.expressionToTokens(normalized)
	}
	if sm := varRefPattern.FindStringSubmatch(normalized); sm != nil {
		return g.expressionToTokens(normalized)
	}
	if sm := parentRefPattern.FindStringSubmatch(normalized); sm != nil {
		return g.expressionToTokens(normalized)
	}
	if strings.HasPrefix(normalized, "module.") || strings.HasPrefix(normalized, "parent.") || strings.HasPrefix(normalized, "var.") {
		return g.expressionToTokens(normalized)
	}

	// Case 2: String with interpolations, e.g., "http://${module.dns.domain}"
	matches := interpolationPattern.FindAllStringSubmatchIndex(normalized, -1)
	if len(matches) == 0 {
		// Just a plain string literal
		return hclwrite.TokensForValue(cty.StringVal(normalized))
	}

	// It's an interpolated string
	var tokens hclwrite.Tokens
	tokens = append(tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenOQuote,
		Bytes: []byte(`"`),
	})

	lastIndex := 0
	for _, match := range matches {
		start, end := match[0], match[1]
		refStart, refEnd := match[2], match[3]

		// Add literal part
		if start > lastIndex {
			tokens = append(tokens, &hclwrite.Token{
				Type:  hclsyntax.TokenStringLit,
				Bytes: []byte(normalized[lastIndex:start]),
			})
		}

		// Add interpolated part
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenTemplateInterp,
			Bytes: []byte(`${`),
		})
		// The reference inside ${...}
		refTokens := g.expressionToTokens(strings.TrimSpace(normalized[refStart:refEnd]))
		tokens = append(tokens, refTokens...)

		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenTemplateSeqEnd,
			Bytes: []byte("}"),
		})

		lastIndex = end
	}

	// Add trailing literal part
	if lastIndex < len(normalized) {
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenStringLit,
			Bytes: []byte(normalized[lastIndex:]),
		})
	}

	tokens = append(tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenCQuote,
		Bytes: []byte(`"`),
	})

	return tokens
}

func (g *Generator) mapToTokens(m map[string]interface{}) (hclwrite.Tokens, error) {
	var tokens hclwrite.Tokens
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}})
	if len(m) > 0 {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		valTokens, err := g.valueToTokens(m[k])
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tokensForObjectKey(k)...)
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}})
		tokens = append(tokens, valTokens...)
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}

	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}})
	return tokens, nil
}

func (g *Generator) sliceToTokens(s []interface{}) (hclwrite.Tokens, error) {
	var tokens hclwrite.Tokens
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}})
	if len(s) > 0 {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}

	for _, v := range s {
		valTokens, err := g.valueToTokens(v)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, valTokens...)
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}})
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}

	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}})
	return tokens, nil
}

func (g *Generator) setVarReference(body *hclwrite.Body, name, refName string) error {
	secretNames := g.getSecretNames()
	prefix := "local"
	if secretNames[refName] {
		prefix = "var"
	}

	tokens := hclwrite.Tokens{
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(prefix),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenDot,
			Bytes: []byte("."),
		},
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(refName),
		},
	}
	body.SetAttributeRaw(name, tokens)
	return nil
}

func setParentOutputRef(body *hclwrite.Body, name, moduleID, outputName string) {
	body.SetAttributeTraversal(name, hcl.Traversal{
		hcl.TraverseRoot{Name: "data"},
		hcl.TraverseAttr{Name: "terraform_remote_state"},
		hcl.TraverseAttr{Name: "env"},
		hcl.TraverseAttr{Name: "outputs"},
		hcl.TraverseAttr{Name: outputName},
	})
}

func (g *Generator) splitByScope(ids []string) (serviceIDs []string, envIDs []string) {
	for _, id := range ids {
		switch g.moduleScopes[id] {
		case scopeService:
			serviceIDs = append(serviceIDs, id)
		case scopeEnv:
			envIDs = append(envIDs, id)
		}
	}
	return serviceIDs, envIDs
}

func (g *Generator) applyAugmentations(m config.Module) config.Module {
	aug, ok := g.iamAugmentations[m.ID]
	if !ok {
		return m
	}

	if m.Inputs == nil {
		m.Inputs = map[string]interface{}{}
	}

	if len(aug.IamPolicy) > 0 {
		if existing, exists := m.Inputs["iam_policy"]; exists {
			if base, ok := existing.(map[string]interface{}); ok {
				if stmts, ok := base["Statement"].([]interface{}); ok && len(stmts) > 0 {
					if more, ok := aug.IamPolicy["Statement"].([]interface{}); ok {
						base["Statement"] = append(stmts, more...)
					}
					m.Inputs["iam_policy"] = base
				} else {
					// replace empty/invalid with generated policy
					m.Inputs["iam_policy"] = aug.IamPolicy
				}
			} else {
				m.Inputs["iam_policy"] = aug.IamPolicy
			}
		} else {
			m.Inputs["iam_policy"] = aug.IamPolicy
		}
	}

	if trusts, exists := m.Inputs["kubernetes_trusts"]; exists {
		if existing, ok := trusts.([]interface{}); ok {
			for _, t := range aug.KubernetesTrusts {
				m.Inputs["kubernetes_trusts"] = append(existing, t)
			}
		}
	} else if len(aug.KubernetesTrusts) > 0 {
		var list []interface{}
		for _, t := range aug.KubernetesTrusts {
			list = append(list, t)
		}
		m.Inputs["kubernetes_trusts"] = list
	}

	return m
}

func (g *Generator) replaceIntrinsicPlaceholders(val string, replaceEnvLayer bool) string {
	// Normalize legacy {foo} style to ${foo} so the rest of the pipeline can treat them as expressions.
	val = normalizeCurlyPlaceholders(val)

	layerName := g.envName
	if g.isService && g.svcCfg != nil {
		layerName = g.svcCfg.Metadata.Name
	}
	accountOrProject := g.envEntry.Account
	envNameReplacement := g.envKey
	accountReplacement := accountOrProject
	regionReplacement := g.envEntry.Region
	if g.useWorkspace {
		envNameReplacement = "${var.environment}"
		accountReplacement = "${var.account_id}"
		regionReplacement = "${var.region}"
	}

	placeholders := map[string]string{
		"account_id": accountReplacement,
		"project_id": accountReplacement,
		"region":     regionReplacement,
	}
	if replaceEnvLayer {
		placeholders["env_name"] = envNameReplacement
		placeholders["layer_name"] = layerName
		placeholders["parent_name"] = layerName
	}
	repl := strings.NewReplacer(buildPlaceholderPairs(placeholders)...)
	return repl.Replace(val)
}

// expressionToTokens renders bare expressions (no surrounding quotes), supporting module, var, and parent references.
func (g *Generator) expressionToTokens(expr string) hclwrite.Tokens {
	normalized := templatePattern.ReplaceAllString(expr, `${$1}`)

	if sm := moduleRefPattern.FindStringSubmatch(normalized); sm != nil {
		return hclwrite.TokensForTraversal(moduleOutputTraversal(sm[1], sm[2]))
	}
	if sm := varRefPattern.FindStringSubmatch(normalized); sm != nil {
		return hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "local"},
			hcl.TraverseAttr{Name: sm[1]},
		})
	}
	if sm := parentRefPattern.FindStringSubmatch(normalized); sm != nil {
		return hclwrite.TokensForTraversal(parentOutputTraversal(sm[1]))
	}
	if strings.HasPrefix(normalized, "module.") {
		parts := strings.SplitN(normalized, ".", 3)
		if len(parts) == 3 {
			return hclwrite.TokensForTraversal(moduleOutputTraversal(parts[1], parts[2]))
		}
	}
	if strings.HasPrefix(normalized, "parent.") {
		name := strings.TrimPrefix(normalized, "parent.")
		if name != "" && !strings.Contains(name, ".") {
			return hclwrite.TokensForTraversal(parentOutputTraversal(name))
		}
	}
	// Fallback to string literal if unknown pattern
	return hclwrite.TokensForValue(cty.StringVal(normalized))
}

func (g *Generator) findModulesByType(moduleType string) map[string]struct{} {
	return g.findModulesByTypes([]string{moduleType})
}

func (g *Generator) findModulesByTypes(moduleTypes []string) map[string]struct{} {
	typeSet := make(map[string]struct{}, len(moduleTypes))
	for _, t := range moduleTypes {
		if t == "" {
			continue
		}
		typeSet[t] = struct{}{}
	}

	matches := map[string]struct{}{}
	for _, m := range g.allModules {
		if _, ok := typeSet[m.Type]; ok {
			matches[m.ID] = struct{}{}
		}
	}
	return matches
}

func (g *Generator) findFirstModuleByType(moduleType string) string {
	for _, m := range g.allModules {
		if m.Type == moduleType {
			return m.ID
		}
	}
	return ""
}

func (g *Generator) writeBaseFiles(useVarRefs bool, workspaceKeyPrefix string) error {
	provider := g.envCfg.Metadata.Provider
	providerRegion := g.envEntry.Region
	account := g.envEntry.Account
	needsHelm := g.envCfg.Providers.Helm
	needsKustomize := g.envCfg.Providers.Kustomize
	needsK8s := g.envCfg.Providers.Kubernetes || needsHelm || needsKustomize
	if g.svcCfg != nil {
		needsHelm = needsHelm || g.svcCfg.Providers.Helm
		needsKustomize = needsKustomize || g.svcCfg.Providers.Kustomize
		needsK8s = needsK8s || g.svcCfg.Providers.Kubernetes || needsHelm || needsKustomize
	}
	cluster, err := g.clusterRefs()
	if err != nil {
		return err
	}

	// Collect locals and secrets
	locals := g.mergedVars
	secretNames := g.getSecretNames()

	var backendKey string
	backendCfg, err := ResolveBackendConfig(provider, g.envCfg, g.envEntry)
	if err != nil {
		return err
	}
	bucket := backendCfg.Bucket
	if useVarRefs {
		backendKey = "terraform.tfstate"
	} else if g.isService {
		backendKey = fmt.Sprintf("service/%s/%s/terraform.tfstate", g.svcCfg.Metadata.Name, g.envKey)
	} else {
		backendKey = fmt.Sprintf("env/%s/%s/terraform.tfstate", g.envCfg.Metadata.Name, g.envKey)
	}

	if bucket == "" {
		return fmt.Errorf("backend bucket is not specified in the configuration")
	}

	if err := writeVersionsTF(g.outDir, bucket, backendKey, backendCfg.Region, provider, backendCfg.BackendType, locals, needsK8s, needsHelm, needsKustomize, backendCfg.Container, backendCfg.ResourceGroup, backendCfg.Profile, useVarRefs, workspaceKeyPrefix); err != nil {
		return fmt.Errorf("failed to write versions.tf: %w", err)
	}

	if err := writeProvidersTF(g.outDir, provider, providerRegion, account, needsK8s, needsHelm, needsKustomize, cluster, useVarRefs); err != nil {
		return fmt.Errorf("failed to write providers.tf: %w", err)
	}

	if err := writeSecretsTF(g.outDir, secretNames); err != nil {
		return fmt.Errorf("failed to write secrets.tf: %w", err)
	}

	// For services, write remote state to access env outputs
	if g.isService {
		envStateKey := fmt.Sprintf("env/%s/%s/terraform.tfstate", g.envCfg.Metadata.Name, g.envKey)
		workspacePrefix := ""
		if useVarRefs {
			envStateKey = "terraform.tfstate"
			workspacePrefix = fmt.Sprintf("env/%s", g.envCfg.Metadata.Name)
		}
		if err := writeRemoteStateTF(g.outDir, backendCfg.BackendType, backendCfg.Bucket, envStateKey, backendCfg.Region, backendCfg.Container, backendCfg.ResourceGroup, backendCfg.Profile, workspacePrefix, useVarRefs); err != nil {
			return fmt.Errorf("failed to write service state.tf: %w", err)
		}
	}

	return nil
}

func (g *Generator) getMergedVars() map[string]interface{} {
	// Precedence: environment vars -> service envRef vars -> CLI --var overrides.
	merged := map[string]interface{}{}
	merged["account_id"] = g.envEntry.Account
	merged["region"] = g.envEntry.Region
	merged["environment"] = g.envKey
	if g.globalLabels == nil {
		merged["global_tags"] = map[string]string{}
	} else {
		merged["global_tags"] = g.globalLabels
	}

	// Env vars
	for k, v := range g.envEntry.Variables {
		merged[k] = parseVarValue(v)
	}

	// Service vars
	if g.isService {
		for k, v := range g.svcEnvEntry.Variables {
			merged[k] = parseVarValue(v)
		}
	}

	// CLI vars (highest precedence)
	for k, v := range g.cliVars {
		merged[k] = parseVarValue(v)
	}
	return merged
}

func (g *Generator) getSecretNames() map[string]bool {
	// Secrets come from env entry and optionally service envRef.
	secrets := map[string]bool{}
	if g.envEntry.Secrets != nil {
		for name := range g.envEntry.Secrets {
			secrets[name] = true
		}
	}
	if g.isService && g.svcEnvEntry.Secrets != nil {
		for name := range g.svcEnvEntry.Secrets {
			secrets[name] = true
		}
	}
	return secrets
}

func (g *Generator) addGlobalTags() {
	envLabels := g.envCfg.Metadata.Labels
	svcLabels := map[string]string{}
	if g.svcCfg != nil {
		svcLabels = g.svcCfg.Metadata.Labels
	}
	g.globalLabels = mergeLabels(envLabels, svcLabels)
	if g.globalLabels == nil {
		g.globalLabels = map[string]string{}
	}
}

func (g *Generator) addDep(modID, depID string) {
	if modID == "" || depID == "" || modID == depID {
		return
	}
	if _, ok := g.moduleDeps[modID]; !ok {
		g.moduleDeps[modID] = map[string]struct{}{}
	}
	g.moduleDeps[modID][depID] = struct{}{}
}

func (g *Generator) collectDepsFromValue(modID string, v interface{}) {
	switch val := v.(type) {
	case string:
		addModuleDep := func(s string) {
			if sm := moduleRefPattern.FindStringSubmatch(s); len(sm) > 1 {
				g.addDep(modID, sm[1])
			}
			for _, match := range moduleRefAnywhere.FindAllStringSubmatch(s, -1) {
				if len(match) > 1 {
					g.addDep(modID, match[1])
				}
			}
		}
		// look inside ${...}
		for _, m := range interpolationPattern.FindAllStringSubmatch(val, -1) {
			if len(m) > 1 {
				addModuleDep(strings.TrimSpace(m[1]))
			}
		}
		addModuleDep(val)
	case []interface{}:
		for _, item := range val {
			g.collectDepsFromValue(modID, item)
		}
	case []string:
		for _, item := range val {
			g.collectDepsFromValue(modID, item)
		}
	case map[string]interface{}:
		for _, item := range val {
			g.collectDepsFromValue(modID, item)
		}
	case map[string]string:
		for _, item := range val {
			g.collectDepsFromValue(modID, item)
		}
	}
}

func (g *Generator) sortedDeps(modID string) []string {
	set, ok := g.moduleDeps[modID]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(set))
	for d := range set {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

func (g *Generator) dependsTokens(deps []string) hclwrite.Tokens {
	var tokens hclwrite.Tokens
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenOBrack, Bytes: []byte{'['}})
	for i, dep := range deps {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
		trav := hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: dep},
		}
		tokens = append(tokens, hclwrite.TokensForTraversal(trav)...)
		if i < len(deps)-1 {
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}})
		}
	}
	if len(deps) > 0 {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte{']'}})
	return tokens
}

func preserveTerraformCache(outDir string) (func() error, error) {
	tfDir := filepath.Join(outDir, ".terraform")
	info, err := os.Stat(tfDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat terraform cache %s: %w", tfDir, err)
	}
	if !info.IsDir() {
		if err := os.Remove(tfDir); err != nil {
			return nil, fmt.Errorf("failed to remove stale terraform cache %s: %w", tfDir, err)
		}
		return nil, nil
	}

	tmpDir, err := os.MkdirTemp("", "pltf-terraform-cache-")
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform cache temp dir: %w", err)
	}
	cached := filepath.Join(tmpDir, ".terraform")
	if err := os.Rename(tfDir, cached); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to relocate terraform cache: %w", err)
	}
	return func() error {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("failed to ensure output dir %s exists: %w", outDir, err)
		}
		target := filepath.Join(outDir, ".terraform")
		if err := os.RemoveAll(target); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clean terraform cache destination %s: %w", target, err)
		}
		if err := os.Rename(cached, target); err != nil {
			return fmt.Errorf("failed to restore terraform cache: %w", err)
		}
		return os.RemoveAll(tmpDir)
	}, nil
}
