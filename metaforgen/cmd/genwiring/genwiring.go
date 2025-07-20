package genwiring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"metaforgen/config"
	"metaforgen/wiringgen"
)

func RunWiringGeneration(configPath, rootModuleName, workflowModulePath, outputDir string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg config.SystemConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if err := wiringgen.GenerateWiringSpec(cfg, workflowModulePath, outputDir+"/wiring"); err != nil {
		return fmt.Errorf("failed to generate wiring spec: %w", err)
	}
	fmt.Println("Wiring spec generated at", filepath.Join(outputDir, "wiring", "specs", "docker.go"))

	// Also generate wiring/main.go
	err = wiringgen.GenerateBlueprintMainFile(
		outputDir+"/wiring",            // directory to write main.go
		rootModuleName,                 // app name
		workflowModulePath,             // workflow spec module path
		rootModuleName+"/wiring/specs", // wiring spec import path
		"Docker",                       // spec variable name
	)
	if err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}
	fmt.Println("main.go generated at", filepath.Join(outputDir, "main.go"))
	if err := wiringgen.GenerateGoMod(rootModuleName+"/wiring", workflowModulePath, filepath.Join(outputDir, "wiring", "go.mod")); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}
	fmt.Println("go.mod generated at", filepath.Join(outputDir, "go.mod"))

	return nil
}
