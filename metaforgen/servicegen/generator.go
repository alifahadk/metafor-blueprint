package servicegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"metaforgen/config"
	"metaforgen/utils"
)

type ServiceData struct {
	PackageName       string
	ServiceName       string
	InterfaceName     string
	ImplName          string
	Constructor       string
	Methods           []MethodData
	Dependencies      []Dependency
	MethodDownstreams map[string][]DownstreamCall
	Imports           []string
}

type MethodData struct {
	Name          string
	SleepDuration int
}

type Dependency struct {
	FieldName     string
	InterfaceName string
	ImportPath    string
}

type DownstreamCall struct {
	FieldName string
	APIName   string
}

func GenerateServices(cfg config.SystemConfig, workflowModulePath, outDir string) error {
	tmpl := template.Must(template.New("service").Parse(ServiceSpecTemplate))
	serverMap := map[string]config.Server{}
	for _, s := range cfg.Servers {
		serverMap[s.Name] = s
	}

	for _, server := range cfg.Servers {
		serviceName := fmt.Sprintf("svc_%s", server.Name)
		serviceDir := filepath.Join(outDir, serviceName)
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			return err
		}

		depMap := map[string]Dependency{}
		methodDownstreams := map[string][]DownstreamCall{}
		var methods []MethodData

		for apiName, apiCfg := range server.APIs {
			methodName := utils.ToTitle(apiName)
			var downstreamCalls []DownstreamCall

			for _, ds := range apiCfg.DownstreamServices {
				targetService := fmt.Sprintf("svc_%s", ds.Target)
				if targetService == "" {
					continue
				}

				importPath := fmt.Sprintf("%s/%s", workflowModulePath, targetService)
				suffix := strings.TrimPrefix(targetService, "svc_")
				fieldName := targetService + utils.ToTitle(suffix)

				if _, ok := depMap[targetService]; !ok && targetService != serviceName {
					depMap[targetService] = Dependency{
						FieldName:     fieldName,
						InterfaceName: fmt.Sprintf("%s.Service%s", targetService, utils.ToTitle(targetService)),
						ImportPath:    importPath,
					}
				}

				downstreamCalls = append(downstreamCalls, DownstreamCall{
					FieldName: fieldName,
					APIName:   utils.ToTitle(ds.API),
				})
			}

			if len(downstreamCalls) > 0 {
				methodDownstreams[methodName] = downstreamCalls
			}

			methods = append(methods, MethodData{
				Name:          methodName,
				SleepDuration: int(apiCfg.ProcessingRate),
			})
		}

		deps := make([]Dependency, 0, len(depMap))
		importSet := map[string]struct{}{}
		for _, d := range depMap {
			deps = append(deps, d)
			importSet[d.ImportPath] = struct{}{}
		}
		var imports []string
		for imp := range importSet {
			imports = append(imports, imp)
		}

		serviceData := ServiceData{
			PackageName:       serviceName,
			ServiceName:       serviceName,
			InterfaceName:     fmt.Sprintf("Service%s", utils.ToTitle(serviceName)),
			ImplName:          fmt.Sprintf("Service%sImpl", utils.ToTitle(serviceName)),
			Constructor:       fmt.Sprintf("NewService%s", utils.ToTitle(serviceName)),
			Methods:           methods,
			Dependencies:      deps,
			MethodDownstreams: methodDownstreams,
			Imports:           imports,
		}

		serviceFile := filepath.Join(serviceDir, serviceName+".go")
		f, err := os.Create(serviceFile)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := tmpl.Execute(f, serviceData); err != nil {
			return err
		}

		fmt.Println("Generated:", serviceFile)
	}

	return nil
}

type ModFileData struct {
	ModuleName string
}

func GenerateGoMod(moduleName, path string) error {
	tmpl := template.Must(template.New("gomod").Parse(GoModTemplate))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := tmpl.Execute(f, ModFileData{
		ModuleName: moduleName,
	}); err != nil {
		return err
	}

	fmt.Println("Generated:", path)
	return nil
}
