package workerpool

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

type serverArgs struct {
	Package         golang.PackageInfo
	Service         *gocode.ServiceInterface
	Iface           *gocode.ServiceInterface
	Name            string
	IfaceName       string
	Imports         *gogen.Imports
	ServerIfaceName string
}

// [Instrument] can be called from the wiring specification to add logging statements to every method in both the server and client side code of Service with `serviceName`.
func Instrument(spec wiring.WiringSpec, serviceName string) {
	// Define the names for the wrapper nodes we are adding to the Blueprint IR
	wrapper_name := serviceName + ".hello.instrument.server"
	client_wrapper_name := serviceName + ".hello.instrument.client"

	// Get the pointer for the serviceName to ensure that the newly defined wrapper IR node will be attached to the node chain of the desired service
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add instrument " + serviceName + " as it is not a pointer. Did you forget to define " + serviceName + "? You can define serviceName using `workflow.Service`")
		return
	}

	// Attach the Hello wrapper node to the server-side node chain of the desired service
	serverNext := ptr.AddDstModifier(spec, wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(wrapper_name, &HelloInstrumentServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, serverNext, err)
		}

		// Instantiate the IRNode
		return newHelloInstrumentServerWrapper(wrapper_name, server)
	})

	// Attach the Hello wrapper node to the client-side node chain of the desired service
	clientNext := ptr.AddSrcModifier(spec, client_wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(client_wrapper_name, &HelloInstrumentClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var client golang.Service
		if err := ns.Get(clientNext, &client); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, clientNext, err)
		}

		return newHelloInstrumentClientWrapper(client_wrapper_name, client)
	})
}

// Blueprint IRNode for representing the wrapper node that instruments every server-side method in the node that gets wrapped
type HelloInstrumentServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloInstrumentServerWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloInstrumentServerWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloInstrumentServerWrapper) String() string {
	return node.Name() + " = HelloInstrumentServerWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.ProvidesInterface
func (node *HelloInstrumentServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func newHelloInstrumentServerWrapper(name string, server ir.IRNode) (*HelloInstrumentServerWrapper, error) {
	serverNode, ok := server.(golang.Service)
	if !ok {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HelloInstrumentServerWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial"

	return node, nil
}

// Implements service.ServiceNode
func (node *HelloInstrumentServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloInstrumentServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	return generateServerInstrumentHandler(builder, iface, node.outputPackage)
}

// Implements golang.Instantiable
func (node *HelloInstrumentServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_TutorialInstrumentServerWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}

func generateServerInstrumentHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Iface:     wrapped,
		Name:      wrapped.BaseName + "_TutorialInstrumentServerWrapper",
		IfaceName: wrapped.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "log", "fmt", "errors")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, wrapped.BaseName+"_TutorialInstrumentServerWrapper"))
	outputFile := filepath.Join(server.Package.Path, wrapped.BaseName+"_TutorialInstrumentServerWrapper.go")
	return gogen.ExecuteTemplateToFile("Tutorial", serverInstrumentTemplate, server, outputFile)
}

var serverInstrumentTemplate = `// Blueprint: Auto-generated by Tutorial Plugin
package {{.Package.ShortName}}

{{.Imports}}


type Job struct {
	Method          func (context.Context) error
	Context        	context.Context
	Done            chan struct{} 
}

func StartWorkerPool(workerCount, queueSize int) chan Job {
	jobQueue := make(chan Job, queueSize)

	for i := 1; i <= workerCount; i++ {
		go func(id int, jobs <-chan Job) { //Read-only channel using <-chan Job instead of chan Job
			for job := range jobs {
				err:=job.Method(job.Context)
				if err!=nil {
					fmt.Println("Error when executing job:",err)
				} else {
                    fmt.Println("Job executed successfully")
                }
				close(job.Done)
			}
		}(i, jobQueue)
	}

	return jobQueue
}
type {{.Name}} struct {
	Service {{.Imports.NameOf .Service.UserType}}
	Queue chan Job
}

func New_{{.Name}}(ctx context.Context, service {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	handler.Queue = StartWorkerPool(4,10)
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Service.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetTypes $f "error"}}) {
	log.Println("Processing {{$f.Name}}")

	done := make(chan struct{})
	job := Job{Method: handler.Service.{{$f.Name}}, Context:{{ArgVars $f "ctx"}}, Done:done}

	select {
	case handler.Queue <- job:
		<-done // Wait for job to complete
		return nil
	default:
		return errors.New("Job count not be executed successfully")
	}
}
{{end}}
`

//Storing this for reference
// return handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})

// Blueprint IRNode for representing the wrapper node that instruments every client-side method in the node that gets wrapped
type HelloInstrumentClientWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloInstrumentClientWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloInstrumentClientWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloInstrumentClientWrapper) String() string {
	return node.Name() + " = HelloInstrumentClientWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.ProvidesInterface
func (node *HelloInstrumentClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func newHelloInstrumentClientWrapper(name string, wrapped ir.IRNode) (*HelloInstrumentClientWrapper, error) {
	serverNode, ok := wrapped.(golang.Service)
	if !ok {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &HelloInstrumentClientWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial"

	return node, nil
}

// Implements service.ServiceNode
func (node *HelloInstrumentClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloInstrumentClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	return generateClientInstrumentHandler(builder, iface, node.outputPackage)
}

// Implements golang.Instantiable
func (node *HelloInstrumentClientWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_TutorialInstrumentClientWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}

func generateClientInstrumentHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Iface:     wrapped,
		Name:      wrapped.BaseName + "_TutorialInstrumentClientWrapper",
		IfaceName: wrapped.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "log")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, wrapped.BaseName+"_TutorialInstrumentClientWrapper"))
	outputFile := filepath.Join(server.Package.Path, wrapped.BaseName+"_TutorialInstrumentClientWrapper.go")
	return gogen.ExecuteTemplateToFile("Tutorial", clientInstrumentTemplate, server, outputFile)
}

var clientInstrumentTemplate = `// Blueprint: Auto-generated by Tutorial Plugin
package {{.Package.ShortName}}

{{.Imports}}




type {{.Name}} struct {
	Client {{.Imports.NameOf .Service.UserType}}
}

func New_{{.Name}}(ctx context.Context, client {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Client = client
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Service.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetTypes $f "error"}}) {
	log.Println("Processing {{$f.Name}}")
	return handler.Client.{{$f.Name}}({{ArgVars $f "ctx"}})
}
{{end}}
`
