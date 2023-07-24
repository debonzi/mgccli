package sdk

import (
	"fmt"
	"regexp"
	"strings"

	"magalu.cloud/core"

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/slices"
)

// Source -> Module -> Resource -> Operation

// Operation

type OpenApiOperation struct {
	key           string
	method        string
	path          *openapi3.PathItem
	operation     *openapi3.Operation
	doc           *openapi3.T
	paramsSchema  *Schema
	configsSchema *Schema
	// TODO: configsMapping map[string]...
	extensionPrefix *string
}

// BEGIN: Descriptor interface:

var openAPIPathArgRegex = regexp.MustCompile("[{](?P<name>[^}]+)[}]")

func getActionName(httpMethod string, pathName string) string {
	name := []string{string(httpMethod)}
	hasArgs := false

	for _, pathEntry := range strings.Split(pathName, "/") {
		match := openAPIPathArgRegex.FindStringSubmatch(pathEntry)
		for i, substr := range match {
			if openAPIPathArgRegex.SubexpNames()[i] == "name" {
				name = append(name, substr)
				hasArgs = true
			}
		}
		if len(match) == 0 && hasArgs {
			name = append(name, pathEntry)
		}
	}

	return strings.Join(name, "-")
}

func (o *OpenApiOperation) Name() string {
	name := getNameExtension(o.extensionPrefix, o.operation.Extensions, "")
	if name == "" {
		name = getActionName(o.method, o.key)
	}
	return name
}

func (o *OpenApiOperation) Version() string {
	return ""
}

func (o *OpenApiOperation) Description() string {
	return o.operation.Description
}

// END: Descriptor interface

// BEGIN: Executor interface:

func addParameters(schema *Schema, parameters openapi3.Parameters, extensionPrefix *string) {
	for _, ref := range parameters {
		parameter := ref.Value
		name := getNameExtension(extensionPrefix, parameter.Schema.Value.Extensions, parameter.Name)

		schema.Properties[name] = parameter.Schema

		if parameter.Required && !slices.Contains(schema.Required, name) {
			schema.Required = append(schema.Required, name)
		}
	}
}

func addRequestBodyParameters(schema *Schema, rbr *openapi3.RequestBodyRef, extensionPrefix *string) {
	if rbr == nil {
		return
	}

	rb := rbr.Value
	mt := rb.Content.Get("application/json")
	if mt == nil {
		return
	}

	content := mt.Schema.Value
	if content == nil {
		return
	}

	for name, ref := range content.Properties {
		parameter := ref.Value
		name = getNameExtension(extensionPrefix, parameter.Extensions, name)

		for {
			_, exists := schema.Properties[name]
			if exists {
				name = "req-" + name
			} else {
				break
			}
		}

		schema.Properties[name] = ref

		if slices.Contains(content.Required, name) && !slices.Contains(schema.Required, name) {
			schema.Required = append(schema.Required, name)
		}
	}
}

func (o *OpenApiOperation) ParametersSchema() *Schema {
	if o.paramsSchema == nil {
		rootSchema := core.NewObjectSchema(map[string]*Schema{}, []string{})

		addParameters(rootSchema, o.path.Parameters, o.extensionPrefix)
		addParameters(rootSchema, o.operation.Parameters, o.extensionPrefix)
		addRequestBodyParameters(rootSchema, o.operation.RequestBody, o.extensionPrefix)

		o.paramsSchema = rootSchema
	}
	return o.paramsSchema
}

func (o *OpenApiOperation) ConfigsSchema() *Schema {
	if o.configsSchema == nil {
		rootSchema := core.NewObjectSchema(map[string]*Schema{}, []string{})
		// TODO: convert and save
		// likely filter by location, like header/cookie are config?
		o.configsSchema = rootSchema
	}
	return o.configsSchema
}

func (o *OpenApiOperation) Execute(parameters map[string]Value, configs map[string]Value) (result Value, err error) {
	// load definitions if not done yet
	parametersSchema := o.ParametersSchema()
	configsSchema := o.ConfigsSchema()

	parameterErr := parametersSchema.VisitJSON(parameters)
	if parameterErr != nil {
		return nil, parameterErr
	}

	// TODO: Validate configs

	fmt.Printf("TODO: execute: %v %v\ninput: p=%v; c=%v\ndefinitions: p=%v; c=%v\n", o.method, o.key, parameters, configs, parametersSchema.Properties, configsSchema)

	return nil, fmt.Errorf("not implemented")
}

var _ Executor = (*OpenApiOperation)(nil)

// END: Executor interface
