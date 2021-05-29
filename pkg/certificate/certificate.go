package certificate

import (
	"haws/pkg/customtags"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/certificatemanager"
)

type Certificate struct {
	domain string
	zoneId string
	san    []string
}

func New(domain string, zoneId string, san []string) *Certificate {
	return &Certificate{
		domain: domain,
		zoneId: zoneId,
		san:    san,
	}
}

func (c *Certificate) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Resources["HugoSslCertificate"] = &certificatemanager.Certificate{
		DomainName: c.domain,
		DomainValidationOptions: []certificatemanager.Certificate_DomainValidationOption{{
			DomainName:   c.domain,
			HostedZoneId: c.zoneId,
		}},
		SubjectAlternativeNames: c.san,
		ValidationMethod:        "DNS",

		Tags: customtags.New(),
	}

	t.Outputs["CertificateArn"] = cloudformation.Output{
		Value:       cloudformation.Ref("HugoSslCertificate"),
		Description: "ARN of certificate created in us-east-1 for the cloudfront distribution",
	}

	return t
}
