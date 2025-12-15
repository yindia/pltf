package generate

import "pltf/pkg/config"

// GenerateEnvironmentTF renders Terraform for a single environment entry in an Environment config.
func GenerateEnvironmentTF(envCfg *config.EnvironmentConfig, embeddedRoot, customRoot, envName, outDir, specDir string, cliVars map[string]string) error {
	g, err := NewGenerator(envCfg, nil, embeddedRoot, customRoot, envName, outDir, specDir, cliVars)
	if err != nil {
		return err
	}
	return g.Generate()
}

// =====================
// Service
// =====================

// GenerateServiceTF renders Terraform for a service envRef entry using its referenced Environment.
func GenerateServiceTF(svcCfg *config.ServiceConfig, envCfg *config.EnvironmentConfig, embeddedRoot, customRoot, envName, outDir, specDir string, cliVars map[string]string) error {
	g, err := NewGenerator(envCfg, svcCfg, embeddedRoot, customRoot, envName, outDir, specDir, cliVars)
	if err != nil {
		return err
	}
	return g.Generate()
}
