package stack

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	cfn "github.com/awslabs/goformation/v4/cloudformation"
)

type Template interface {
	Build() *cfn.Template
	GetStackName() *string
	GetRegion() *string
	GetExportName(string) string
	GetParameters() []*cloudformation.Parameter
	GetDryRunOutputs() map[string]string
}

type TemplateComponent struct {
	Params        []*cloudformation.Parameter
	Parameters    map[string]cfn.Parameter
	Resources     map[string]*cfn.Resource
	Outputs       map[string]cfn.Output
	DryRunOutputs map[string]string
}

func NewTemplate() TemplateComponent {
	b := TemplateComponent{
		Params:        make([]*cloudformation.Parameter, 0),
		Parameters:    make(map[string]cfn.Parameter, 0),
		Resources:     make(map[string]*cfn.Resource, 0),
		Outputs:       make(map[string]cfn.Output, 0),
		DryRunOutputs: make(map[string]string, 0),
	}

	return b
}

func (t *TemplateComponent) AddParameter(name string, parameter cfn.Parameter, def string) {
	t.Parameters[name] = parameter
	t.Params = append(t.Params, &cloudformation.Parameter{
		ParameterKey:   &name,
		ParameterValue: &def,
	})
}

func (t *TemplateComponent) AddResource(name string, resource cfn.Resource) {
	t.Resources[name] = &resource
}

func (t *TemplateComponent) AddOutput(name string, output cfn.Output, dryRunValue string) {
	t.Outputs[name] = output
	t.DryRunOutputs[name] = dryRunValue
}

func (t *TemplateComponent) GetParameters() []*cloudformation.Parameter {
	return t.Params
}

func (t *TemplateComponent) Build() *cfn.Template {
	tp := cfn.NewTemplate()
	for paramName, ParamDef := range t.Parameters {
		tp.Parameters[paramName] = ParamDef
	}

	for resName, resDef := range t.Resources {
		tp.Resources[resName] = *resDef
	}

	for outName, outDef := range t.Outputs {
		tp.Outputs[outName] = outDef
	}
	return tp
}

func (t *TemplateComponent) GetDryRunOutputs() map[string]string {
	return t.DryRunOutputs
}
