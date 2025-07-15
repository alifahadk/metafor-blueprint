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
)
require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240405152959-f078915d2306 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/otiai10/copy v1.14.0 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/sdk v1.26.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/tools v0.20.0 // indirect
)`
const DockerSpecTemplate = `package specs

import (
	"strings"
	"strconv"
	"fmt"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/retries"
	"github.com/blueprint-uservices/blueprint/plugins/timeouts"
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

	applyLoggerDefaults := func(service_name string,retry int64, timeout string,threadCount, queueSize uint) string {
		procName := strings.ReplaceAll(service_name, "svc", "process")
		cntrName := strings.ReplaceAll(service_name, "svc", "container")
		if retry>0{
			retries.AddRetries(spec, service_name, retry)
		}
		if intVal, err := strconv.Atoi(timeout); intVal > 0 && err == nil {
			timeouts.Add(spec, service_name, timeout+"s")
		}
		if threadCount==0 || queueSize==0 {

			panic(fmt.Errorf("Invalid values for threadCount / queueSize!"))
		}
		workerpool.Instrument(spec, service_name,threadCount,queueSize)
		http.Deploy(spec, service_name)
		goproc.CreateProcess(spec, procName, service_name)
		return linuxcontainer.CreateContainer(spec, cntrName, procName)
	}

{{range .Services}}
	{{.VarName}} := workflow.Service[*{{.Package}}.{{.ServiceID}}](spec, "{{.VarName}}"
	{{- range .Dependencies}}, "{{.}}"{{- end}})
	containers = append(containers, applyLoggerDefaults({{.VarName}},{{.Retry}},"{{.Timeout}}",{{.ThreadCount}},{{.QueueSize}}))
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
