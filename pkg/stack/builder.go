package stack

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	cfn "github.com/awslabs/goformation/v4/cloudformation"
)

type Template interface {
	Build() *cfn.Template
	GetStackName() *string
	GetRegion() *string
	GetOutputName(string) string
	GetParameters() []*cloudformation.Parameter
	DryRunOutputs() map[string]string
}

type Parameter func() *cloudformation.Parameter

func WithParameter(key string, value string) Parameter {
	return func() *cloudformation.Parameter {
		return &cloudformation.Parameter{
			ParameterKey:   &key,
			ParameterValue: &value,
		}
	}
}

type TemplateFactory struct {
	Params []*cloudformation.Parameter
}

func NewTemplate(params ...Parameter) TemplateFactory {
	b := TemplateFactory{
		make([]*cloudformation.Parameter, 0),
	}

	for _, p := range params {
		b.Params = append(b.Params, p())
	}

	return b
}
