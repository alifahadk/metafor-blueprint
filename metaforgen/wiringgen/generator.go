package wiringgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"metaforgen/config"
	"metaforgen/utils"
)

type DockerSpecData struct {
	Services   []DockerService
	RootModule string
}

type DockerService struct {
	VarName            string
	ServiceID          string
	Package            string
	ProcessID          string
	Container          string
	WorkflowModulePath string
	Dependencies       []string
	Retry              int
	Timeout            int
	QueueSize          uint
	ThreadCount        uint
}

type MainTemplateData struct {
	AppName       string
	WorkflowPath  string
	SpecImport    string
	SpecReference string
}

func GenerateWiringSpec(cfg config.SystemConfig, workflowModulePath, outputDir string) error {
	var services []DockerService
	depGraph := utils.ExtractDependencies(cfg)
	ServerConfig := utils.ExtractServerConfig(cfg)

	for _, srv := range cfg.Servers {
		svcID := fmt.Sprintf("ServiceSvc%sImpl", srv.Name)
		pkgName := fmt.Sprintf("svc%s", srv.Name)
		varName := fmt.Sprintf("svc%s", srv.Name)
		process := strings.ReplaceAll(varName, "svc", "process")
		container := strings.ReplaceAll(varName, "svc", "container")

		dependencies := depGraph[varName]
		retry, exists := ServerConfig[varName]["retry"]
		if !exists {
			retry = 0
		}
		timeout, exists := ServerConfig[varName]["timeout"]
		if !exists {
			timeout = 0
		}
		threadPool, exists := ServerConfig[varName]["threadpool"]
		if !exists {
			timeout = 0
		}
		queueSize, exists := ServerConfig[varName]["queue_size"]
		if !exists {
			timeout = 0
		}
		services = append(services, DockerService{
			VarName:            varName,
			ServiceID:          svcID,
			Package:            pkgName,
			ProcessID:          process,
			Container:          container,
			WorkflowModulePath: workflowModulePath,
			Dependencies:       dependencies,
			Retry:              retry,
			Timeout:            timeout,
			ThreadCount:        uint(threadPool),
			QueueSize:          uint(queueSize),
		})
	}

	err := os.MkdirAll(filepath.Join(outputDir, "specs"), 0755)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("docker").Parse(DockerSpecTemplate))
	specFile := filepath.Join(outputDir, "specs", "docker.go")

	f, err := os.Create(specFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, DockerSpecData{Services: services, RootModule: filepath.Dir(workflowModulePath)})
}

func GenerateBlueprintMainFile(outputDir, appName, workflowModulePath, specImportPath, specRef string) error {
	data := MainTemplateData{
		AppName:       appName,
		WorkflowPath:  workflowModulePath,
		SpecImport:    specImportPath,
		SpecReference: specRef,
	}

	tmpl := template.Must(template.New("main").Parse(blueprintMainTemplate))

	mainFilePath := filepath.Join(outputDir, "main.go")
	f, err := os.Create(mainFilePath)
	if err != nil {
		return fmt.Errorf("could not create main.go: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

type ModFileData struct {
	ModuleName         string
	WorkflowModulePath string
	WorkflowModuleName string
	RootModule         string
}

func GenerateGoMod(moduleName, workflowModulePath, path string) error {
	tmpl := template.Must(template.New("gomod").Parse(GoModTemplate))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := tmpl.Execute(f, ModFileData{
		ModuleName:         moduleName,
		WorkflowModulePath: workflowModulePath,
		WorkflowModuleName: filepath.Base(workflowModulePath),
		RootModule:         filepath.Dir(workflowModulePath),
	}); err != nil {
		return err
	}

	fmt.Println("Generated:", path, workflowModulePath)
	return nil
}
