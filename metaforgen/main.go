package main

import (
	"flag"
	"fmt"
	"metaforgen/cmd/genservices"
	"metaforgen/cmd/genwiring"
	"os"
)

func main() {
	// Define and parse the command-line flag
	configPath := flag.String("config", "../config.json", "Path to the configuration JSON file")
	flag.Parse()

	// Validate the config path
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		panic(fmt.Errorf("config file not found: %s", *configPath))
	}

	workflowModuleName := "workflow"
	outputDir := "output"
	rootModuleName := "metaforsim"
	fmt.Println("Starting service generation...")
	if err := genservices.RunServiceGeneration(rootModuleName+"/"+workflowModuleName, *configPath, workflowModuleName, outputDir); err != nil {
		panic(fmt.Errorf("service generation failed: %w", err))
	}
	fmt.Println("Service generation complete.")

	fmt.Println("Starting wiring spec generation...")
	if err := genwiring.RunWiringGeneration(*configPath, rootModuleName, rootModuleName+"/"+workflowModuleName, outputDir); err != nil {
		panic(fmt.Errorf("wiring generation failed: %w", err))
	}

}
