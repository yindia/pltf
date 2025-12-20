package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"pltf/pkg/config"
	"pltf/pkg/generate"
)

func autoValidate(file, env string) error {
	return autoValidateWithOutput(os.Stdout, file, env)
}

func autoValidateWithOutput(out io.Writer, file, env string) error {
	var (
		envName string
		envCfg  *config.EnvironmentConfig
		svcCfg  *config.ServiceConfig
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
		envName, err = selectEnvName(kind, env, cfg, nil)
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
		envName, err = selectEnvName(kind, env, envConfig, svc)
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

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment or Service)", file)
	}

	// Run lint suggestions alongside validation.
	switch kind {
	case "Environment":
		printLintSuggestions(out, kind, envName, nil)
	case "Service":
		printLintSuggestions(out, kind, envName, nil)
	}
	return nil
}

func autoGenerate(file, env, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := parseVarFlags(vars)
	if err != nil {
		return err
	}

	embeddedRoot, customRoot, err := resolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}

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
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "env", envName)
		}
		out = filepath.Clean(out)

		if err := generate.GenerateEnvironmentTF(envCfg, embeddedRoot, customRoot, envName, out, specDir, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Environment Terraform for %q (env=%s) into %s\n",
			envCfg.Metadata.Name, envName, out)
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "env", envName)
		}
		out = filepath.Clean(out)

		if err := generate.GenerateServiceTF(svcCfg, envCfg, embeddedRoot, customRoot, envName, out, specDir, cliVars); err != nil {
			return err
		}
		fmt.Printf("Generated Service Terraform for %q (env=%s) into %s\n",
			svcCfg.Metadata.Name, envName, out)
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment or Service)", file)
	}
}

// autoGenerateQuiet renders Terraform without printing status messages. Used by graph command to keep DOT output clean.
func autoGenerateQuiet(file, env, modulesRoot, out string, vars []string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	specDir := filepath.Dir(absFile)

	kind, err := config.DetectKind(file)
	if err != nil {
		return err
	}

	cliVars, err := parseVarFlags(vars)
	if err != nil {
		return err
	}

	embeddedRoot, customRoot, err := resolveModuleRoots(modulesRoot)
	if err != nil {
		return err
	}

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
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, "env", envName)
		}
		out = filepath.Clean(out)

		if err := generate.GenerateEnvironmentTF(envCfg, embeddedRoot, customRoot, envName, out, specDir, cliVars); err != nil {
			return err
		}
		return nil

	case "Service":
		svcCfg, envCfg, err := config.LoadService(file)
		if err != nil {
			return err
		}
		envName, err := selectEnvName(kind, env, envCfg, svcCfg)
		if err != nil {
			return err
		}
		if out == "" {
			out = filepath.Join(".pltf", envCfg.Metadata.Name, svcCfg.Metadata.Name, "env", envName)
		}
		out = filepath.Clean(out)

		if err := generate.GenerateServiceTF(svcCfg, envCfg, embeddedRoot, customRoot, envName, out, specDir, cliVars); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown or missing kind in %s (expected Environment or Service)", file)
	}
}
