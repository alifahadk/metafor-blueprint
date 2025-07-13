package wiringgen

const GoModTemplate = `module {{.ModuleName}}

go 1.22

require {{.WorkflowModulePath}} v0.0.0
replace {{.WorkflowModulePath}} => ../{{.WorkflowModuleName}}

require {{.RootModule}}/workerpool v0.0.0
replace {{.RootModule}}/workerpool => ../workerpool
require (
	github.com/blueprint-uservices/blueprint/blueprint v0.0.0-20240124230554-8949221e29cc
	github.com/blueprint-uservices/blueprint/plugins v0.0.0-20240124230554-8949221e29cc
)`
const DockerSpecTemplate = `package specs

import (

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"{{.RootModule}}/workerpool"
{{- range .Services}}
	"{{.WorkflowModulePath}}/{{.Package}}"
{{- end}}
)

var Docker = cmdbuilder.SpecOption{
	Name:        "docker",
	Description: "Auto-generated wiring spec based on service config.",
	Build:       makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string
{{range .Services}}
	{{.VarName}} := workflow.Service[*{{.Package}}.{{.ServiceID}}](spec, "{{.VarName}}")
	tutorial.Instrument(spec, {{.VarName}})
	http.Deploy(spec, {{.VarName}})
	goproc.CreateProcess(spec, "{{.ProcessID}}", {{.VarName}})
	cntr := linuxcontainer.CreateContainer(spec, "{{.Container}}", "{{.ProcessID}}")
	containers = append(containers, cntr)
{{end}}
	return containers, nil
}
`

const blueprintMainTemplate = `

// Package main provides the {{.AppName}} application.
package main

import (
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"{{.SpecImport}}"
)

func main() {
	// Configure the location of our workflow spec
	workflowspec.AddModule("{{.WorkflowPath}}")

	// Build a supported wiring spec
	name := "{{.AppName}}"
	cmdbuilder.MakeAndExecute(
		name,
		specs.{{.SpecReference}},
	)
}
`
