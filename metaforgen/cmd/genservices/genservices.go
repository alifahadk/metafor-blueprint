package genservices

import (
	"fmt"
	"os"
	"path/filepath"

	"metaforgen/config"
	"metaforgen/servicegen"
)

func RunServiceGeneration(workflowModulePath, configPath, modDir, outDir string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(outDir, modDir), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := servicegen.GenerateServices(cfg, workflowModulePath, filepath.Join(outDir, modDir)); err != nil {
		return fmt.Errorf("failed to generate services: %w", err)
	}

	if err := servicegen.GenerateGoMod(workflowModulePath, filepath.Join(filepath.Join(outDir, modDir), "go.mod")); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	return nil
}
