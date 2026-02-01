package clihelper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"pltf/pkg/config"
	"pltf/pkg/generate"
	"pltf/pkg/validate"
)

func AutoValidate(file, env, modulesRoot string) error {
	return AutoValidateWithOutput(os.Stdout, file, env, modulesRoot)
}

func AutoValidateWithOutput(out io.Writer, file, env, modulesRoot string) error {
	var (
		envCfg *config.EnvironmentConfig
		svcCfg *config.ServiceConfig
		// envName string
	)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	switch kind {
	case "Environment":
		cfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		_, err = SelectEnvName(kind, env, cfg, nil)
		if err != nil {
			return err
		}
		envCfg = cfg
		fmt.Fprintf(out, "Environment %q is valid (provider=%s, org=%s)\n",
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
			envCfg.Metadata.Org,
		)

	case "Service":
		svc, envConfig, err := config.LoadService(file)
		if err != nil {
			return err
		}
		_, err = SelectEnvName(kind, env, envConfig, svc)
		if err != nil {
			return err
		}
		svcCfg = svc
		envCfg = envConfig
		fmt.Fprintf(out, "Service %q is valid and uses Environment %q (provider=%s)\n",
			svcCfg.Metadata.Name,
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
		)

	case "Stack":
		stackCfg, err := config.LoadStack(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Stack %q is valid\n", stackCfg.Metadata.Name)
		embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
		if err != nil {
			return err
		}
		validator := validate.NewModuleValidator(embeddedRoot, customRoot)
		if err := validator.ValidateDuplicateOutputsForModules(stackCfg.Modules); err != nil {
			return err
		}
		if err := validator.ValidateClusterModulesForModules(stackCfg.Modules); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}
	validator := validate.NewModuleValidator(embeddedRoot, customRoot)
	if err := validator.ValidateCustomModules(envCfg, svcCfg); err != nil {
		return err
	}
	if err := validate.ValidateProviderSupport(envCfg, svcCfg); err != nil {
		return err
	}
	if err := validator.ValidateDuplicateOutputs(envCfg, svcCfg); err != nil {
		return err
	}
	if err := validator.ValidateClusterModules(envCfg, svcCfg); err != nil {
		return err
	}

	// Run lint suggestions alongside validation.
	return nil
}

func AutoValidateWithScan(out io.Writer, file, env, modules string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	embeddedRoot, customRoot, err := ResolveModuleRoots(modules)
	if err != nil {
		return err
	}
	validator := validate.NewModuleValidator(embeddedRoot, customRoot)

	outDir, err := os.MkdirTemp("", "pltf-validate-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(outDir)

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, nil); err != nil {
			return err
		}
		if err := validator.ValidateClusterModules(envCfg, nil); err != nil {
			return err
		}
		if err := validate.ValidateProviderSupport(envCfg, nil); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Environment %q is valid (provider=%s, org=%s)\n",
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
			envCfg.Metadata.Org,
		)
		if err := generate.ExportEnvironmentWorkspaceForEnv(envCfg, embeddedRoot, customRoot, outDir, specDir, envName, nil); err != nil {
			return err
		}

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, svcCfg); err != nil {
			return err
		}
		if err := validator.ValidateClusterModules(envCfg, svcCfg); err != nil {
			return err
		}
		if err := validate.ValidateProviderSupport(envCfg, svcCfg); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Service %q is valid and uses Environment %q (provider=%s)\n",
			svcCfg.Metadata.Name,
			envCfg.Metadata.Name,
			envCfg.Metadata.Provider,
		)
		if err := generate.ExportServiceWorkspaceForEnv(svcCfg, envCfg, embeddedRoot, customRoot, outDir, specDir, envName, nil); err != nil {
			return err
		}

	case "Stack":
		stackCfg, err := config.LoadStack(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Stack %q is valid (scan not supported for Stack specs)\n", stackCfg.Metadata.Name)
		if err := validator.ValidateDuplicateOutputsForModules(stackCfg.Modules); err != nil {
			return err
		}
		if err := validator.ValidateClusterModulesForModules(stackCfg.Modules); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}

	if _, err := RunTfsecScan(outDir); err != nil {
		return fmt.Errorf("tfsec scan failed: %w", err)
	}
	return nil
}

func AutoGenerate(file, env, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := ParseVarFlags(vars)
	if err != nil {
		return err
	}
	cliVars = MergeVarMaps(ParseVarEnv(), cliVars)

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}
	validator := validate.NewModuleValidator(embeddedRoot, customRoot)

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, nil); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportEnvironmentWorkspaceForEnv(envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Environment workspace for %q (env=%s) into %s\n",
			envCfg.Metadata.Name, envName, out)
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, svcCfg); err != nil {
			return err
		}
		if err := validator.ValidateClusterModules(envCfg, svcCfg); err != nil {
			return err
		}
		if err := validator.ValidateDuplicateOutputs(envCfg, svcCfg); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportServiceWorkspaceForEnv(svcCfg, envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Service workspace for %q (env=%s) into %s\n",
			svcCfg.Metadata.Name, envName, out)
		return nil

	case "Stack":
		return fmt.Errorf("stack specs cannot be generated directly; reference them from Environment or Service specs")

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}
}

func AutoGenerateAll(file, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := ParseVarFlags(vars)
	if err != nil {
		return err
	}
	cliVars = MergeVarMaps(ParseVarEnv(), cliVars)

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}
	validator := validate.NewModuleValidator(embeddedRoot, customRoot)

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, nil); err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportEnvironmentWorkspace(envCfg, embeddedRoot, customRoot, out, specDir, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Environment workspace for %q into %s\n", envCfg.Metadata.Name, out)
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, svcCfg); err != nil {
			return err
		}
		if err := validator.ValidateClusterModules(envCfg, svcCfg); err != nil {
			return err
		}
		if err := validator.ValidateDuplicateOutputs(envCfg, svcCfg); err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportServiceWorkspace(svcCfg, envCfg, embeddedRoot, customRoot, out, specDir, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Service workspace for %q into %s\n", svcCfg.Metadata.Name, out)
		return nil

	case "Stack":
		return fmt.Errorf("stack specs cannot be generated directly; reference them from Environment or Service specs")

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment, Service, or Stack)", file)
	}
}

// autoGenerateQuiet renders Terraform without printing status messages. Used by graph command to keep DOT output clean.
func AutoGenerateQuiet(file, env, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := ParseVarFlags(vars)
	if err != nil {
		return err
	}
	cliVars = MergeVarMaps(ParseVarEnv(), cliVars)

	embeddedRoot, customRoot, err := ResolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}
	validator := validate.NewModuleValidator(embeddedRoot, customRoot)

	switch kind {
	case "Environment":
		envCfg, err := config.LoadEnvironmentConfig(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, nil); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, nil)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportEnvironmentWorkspaceForEnv(envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		if err := validator.ValidateCustomModules(envCfg, svcCfg); err != nil {
			return err
		}
		envName, err := SelectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "workspace")
		}
		out = filepath.Clean(out)

		if err := generate.ExportServiceWorkspaceForEnv(svcCfg, envCfg, embeddedRoot, customRoot, out, specDir, envName, cliVars); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment or Service)", file)
	}
}
