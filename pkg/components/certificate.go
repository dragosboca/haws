package components

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/components/resources/customtags"
	"github.com/dragosboca/haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/certificatemanager"
)

type Certificate struct {
	stack.TemplateComponent
	Prefix string
}

type CertificateInput struct {
	Prefix string
	Region string
	Domain string
	ZoneId string
}

func NewCertificate(c *CertificateInput) *Certificate {
	certificate := &Certificate{
		Prefix:            c.Prefix,
		TemplateComponent: stack.NewTemplate("us-east-1"),
	}

	certificate.AddParameter("Domain", cloudformation.Parameter{
		Type:        "String",
		Description: "Domain for which we generate the certificate",
	}, c.Domain)

	certificate.AddParameter("ZoneId", cloudformation.Parameter{
		Type:        "String",
		Description: "The Route53 zone used for domain validation",
	}, c.ZoneId)

	certificate.AddResource("HugoSslCertificate", &certificatemanager.Certificate{
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
	})

	certificate.AddOutput("Arn", cloudformation.Output{
		Value:       cloudformation.Ref("HugoSslCertificate"),
		Description: "ARN of certificate created in us-east-1 for the cloudfront distribution",
		Export: &cloudformation.Export{
			Name: certificate.GetExportName("Arn"),
		},
	}, "arn:aws:acm:us-east-1:123456789012:certificate/123456789012-1234-1234-1234-12345678")

	return certificate
}

func (c *Certificate) GetExportName(output string) string {
	return fmt.Sprintf("HawsCertificate%s%s", output, strings.Title(c.Prefix))
}

func (c *Certificate) GetStackName() *string {
	stackName := fmt.Sprintf("%s-certificate", c.Prefix)
	return &stackName
}
