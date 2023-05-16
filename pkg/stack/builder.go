package stack

import (
	cfn "github.com/awslabs/goformation/v4/cloudformation"
)

type Template interface {
	Build() *cfn.Template
	GetStackName() string
	GetRegion() string
	GetExportName(string) string
	GetDryRunOutputs() map[string]string
}

type TemplateComponent struct {
	Region        string
	Parameters    map[string]cfn.Parameter
	Resources     map[string]cfn.Resource
	Outputs       map[string]cfn.Output
	DryRunOutputs map[string]string
}

func NewTemplate(region string) TemplateComponent {
	b := TemplateComponent{
		Parameters:    make(map[string]cfn.Parameter, 0),
		Resources:     make(map[string]cfn.Resource, 0),
		Outputs:       make(map[string]cfn.Output, 0),
		DryRunOutputs: make(map[string]string, 0),
		Region:        region,
	}

	return b
}

func (t *TemplateComponent) AddParameter(name string, parameter cfn.Parameter) {
	t.Parameters[name] = parameter
}

func (t *TemplateComponent) AddResource(name string, resource cfn.Resource) {
	t.Resources[name] = resource
}

func (t *TemplateComponent) AddOutput(name string, output cfn.Output, dryRunValue string) {
	t.Outputs[name] = output
	t.DryRunOutputs[name] = dryRunValue
}

func (t *TemplateComponent) Build() *cfn.Template {
	tp := cfn.NewTemplate()
	for paramName, ParamDef := range t.Parameters {
		tp.Parameters[paramName] = ParamDef
	}

	for resName, resDef := range t.Resources {
		tp.Resources[resName] = resDef
	}

	for outName, outDef := range t.Outputs {
		tp.Outputs[outName] = outDef
	}
	return tp
}

func (t *TemplateComponent) GetDryRunOutputs() map[string]string {
	return t.DryRunOutputs
}

func (t *TemplateComponent) GetRegion() string {
	return t.Region
}
