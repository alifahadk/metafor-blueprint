package main

import (
	"fmt"
	"metaforgen/cmd/genservices"
	"metaforgen/cmd/genwiring"
)

func main() {
	configPath := "config.json"
	workflowModuleName := "workflow"
	outputDir := "output"
	rootModuleName := "metaforsim"
	fmt.Println("Starting service generation...")
	if err := genservices.RunServiceGeneration(rootModuleName+"/"+workflowModuleName, configPath, workflowModuleName, outputDir); err != nil {
		panic(fmt.Errorf("service generation failed: %w", err))
	}
	fmt.Println("Service generation complete.")

	fmt.Println("Starting wiring spec generation...")
	if err := genwiring.RunWiringGeneration(configPath, rootModuleName, rootModuleName+"/"+workflowModuleName, outputDir); err != nil {
		panic(fmt.Errorf("wiring generation failed: %w", err))
	}

}
