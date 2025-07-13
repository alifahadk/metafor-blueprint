package servicegen

const GoModTemplate = `module {{.ModuleName}}

go 1.22

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240118214956-4c7cf2978ee5
	golang.org/x/text v0.13.0
)
`

const ServiceSpecTemplate = `package {{.PackageName}}

import (
	"context"
	"fmt"
	"time"
{{- range .Imports}}
	"{{.}}"
{{- end}}
)

type {{.InterfaceName}} interface {
{{- range .Methods}}
	{{.Name}}(ctx context.Context) error
{{- end}}
}

type {{.ImplName}} struct {
{{- range .Dependencies}}
	{{.FieldName}} {{.InterfaceName}}
{{- end}}
}

func {{.Constructor}}(ctx context.Context{{range .Dependencies}}, {{.FieldName}} {{.InterfaceName}}{{end}}) (*{{.ImplName}}, error) {
	return &{{.ImplName}}{
		{{- range .Dependencies}}
		{{.FieldName}}: {{.FieldName}},
		{{- end}}
	}, nil
}

{{range .Methods}}
func (s *{{$.ImplName}}) {{.Name}}(ctx context.Context) error {
	fmt.Println("{{.Name}} called on {{$.ServiceName}}")
	time.Sleep({{.SleepDuration}} * time.Second)
	{{- with index $.MethodDownstreams .Name}}
	{{- range .}}
	if err := s.{{.FieldName}}.{{.APIName}}(ctx); err != nil {
		return err
	}
	{{- end}}
	{{- end}}
	return nil
}
{{end}}`
