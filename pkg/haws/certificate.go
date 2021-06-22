package haws

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/stack"

	cfn "github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/certificatemanager"
)

type Certificate struct {
	*Haws
	stack.TemplateFactory
	region string
}

func NewCertificate(h *Haws) *Certificate {
	return &Certificate{
		h,
		stack.NewTemplate(
			stack.WithParameter("Domain", h.Domain),
			stack.WithParameter("ZoneId", h.ZoneId),
		),
		"us-east-1",
	}
}

func (c *Certificate) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Parameters["Domain"] = cloudformation.Parameter{
		Type:        "String",
		Description: "Domain for which we generate the certificate",
	}

	t.Parameters["ZoneId"] = cloudformation.Parameter{
		Type:        "String",
		Description: "The Route53 zone used for domain validation",
	}

	t.Resources["HugoSslCertificate"] = &certificatemanager.Certificate{
		DomainName: cloudformation.Ref("Domain"),
		DomainValidationOptions: []certificatemanager.Certificate_DomainValidationOption{{
			DomainName:   cloudformation.Ref("Domain"),
			HostedZoneId: cloudformation.Ref("ZoneId"),
		}},
		SubjectAlternativeNames: []string{
			cloudformation.Ref("Domain"),
		},
		ValidationMethod: "DNS",

		Tags: customtags.New(),
	}

	t.Outputs[c.GetOutputName("Arn")] = cloudformation.Output{
		Value:       cloudformation.Ref("HugoSslCertificate"),
		Description: "ARN of certificate created in us-east-1 for the cloudfront distribution",
	}

	return t
}

func (c *Certificate) GetOutputName(output string) string {
	return fmt.Sprintf("HawsCertificate%s%s", output, strings.Title(c.Prefix))
}

func (c *Certificate) GetStackName() *string {
	stackName := fmt.Sprintf("%sCertificate", c.Prefix)
	return &stackName
}

func (c *Certificate) GetRegion() *string {
	return &c.Region
}

func (c *Certificate) GetParameters() []*cfn.Parameter {
	return c.Params
}

func (c *Certificate) DryRunOutputs() map[string]string {
	ret := make(map[string]string)
	ret[c.GetOutputName("Arn")] = "arn:aws:acm:us-east-1:123456789012:certificate/123456789012-1234-1234-1234-12345678"
	return ret
}
