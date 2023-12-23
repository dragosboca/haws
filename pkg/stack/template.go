package stack

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	cfn "github.com/awslabs/goformation/v4/cloudformation"
)

// Template is an interface that defines the methods that a template must implement
type Template interface {
	Build() *cfn.Template
	GetStackName() *string
	GetRegion() string
	GetExportName(string) string
	GetParameters() []*cloudformation.Parameter
	GetDryRunOutputs() map[string]string
	SetParameterValue(string, string) error
}

// TemplateComponent is a struct that implements the Template interface
// It is used to build a CloudFormation template
type TemplateComponent struct {
	Region        string
	Params        []*cloudformation.Parameter
	Parameters    map[string]cfn.Parameter
	Resources     map[string]*cfn.Resource
	Outputs       map[string]cfn.Output
	DryRunOutputs map[string]string
}

func NewTemplate(region string) TemplateComponent {
	b := TemplateComponent{
		Params:        make([]*cloudformation.Parameter, 0),
		Parameters:    make(map[string]cfn.Parameter),
		Resources:     make(map[string]*cfn.Resource),
		Outputs:       make(map[string]cfn.Output),
		DryRunOutputs: make(map[string]string),
		Region:        region,
	}

	return b
}

// AddParameter adds a parameter to the template
// param: name - the name of the parameter
// param: parameter - the parameter definition
// param: def - the default value of the parameter
func (t *TemplateComponent) AddParameter(name string, parameter cfn.Parameter, def string) {
	t.Parameters[name] = parameter
	t.Params = append(t.Params, &cloudformation.Parameter{
		ParameterKey:   &name,
		ParameterValue: &def,
	})
}

// AddResource adds a resource to the template
// param: name - the name of the resource
// param: resource - the resource definition
func (t *TemplateComponent) AddResource(name string, resource cfn.Resource) {
	t.Resources[name] = &resource
}

// AddOutput adds an output to the template
// param: name - the name of the output
// param: output - the output definition
// param: dryRunValue - the value of the output when running in dry-run mode
func (t *TemplateComponent) AddOutput(name string, output cfn.Output, dryRunValue string) {
	t.Outputs[name] = output
	t.DryRunOutputs[name] = dryRunValue
}

// GetParameters returns the parameters of the template
// return: []*cloudformation.Parameter - the parameters of the template
func (t *TemplateComponent) GetParameters() []*cloudformation.Parameter {
	return t.Params
}

// Build builds the template
// return: *cfn.Template - the template
func (t *TemplateComponent) Build() *cfn.Template {
	tp := cfn.NewTemplate()
	for paramName, paramDef := range t.Parameters {
		tp.Parameters[paramName] = paramDef
	}

	for resName, resDef := range t.Resources {
		tp.Resources[resName] = *resDef
	}

	for outName, outDef := range t.Outputs {
		tp.Outputs[outName] = outDef
	}
	return tp
}

// GetDryRunOutputs returns the outputs of the template when running in dry-run mode
// return: map[string]string - the outputs of the template when running in dry-run mode
func (t *TemplateComponent) GetDryRunOutputs() map[string]string {
	return t.DryRunOutputs
}

// GetRegion returns the region of the template
// return: string - the region of the template
func (t *TemplateComponent) GetRegion() string {
	return t.Region
}

func (t *TemplateComponent) SetParameterValue(name string, value string) error {
	if param, ok := t.Parameters[name]; ok {
		t.Parameters[name] = param
		for _, p := range t.Params {
			if *p.ParameterKey == name {
				p.ParameterValue = &value
			}
		}
		return nil
	}
	return fmt.Errorf("parameter %s not found", name)
}
