package certificate

import (
	"fmt"
	"haws/pkg/customtags"
	"haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/certificatemanager"
)

type Certificate struct {
	domain string
	zoneId string
	san    []string
	prefix string
}

func New(prefix string, domain string, zoneId string, san []string) *Certificate {
	return &Certificate{
		domain: domain,
		zoneId: zoneId,
		san:    san,
		prefix: prefix,
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

func (c *Certificate) Deploy() (stack.Output, error) {
	certStackName := fmt.Sprintf("%sCertificate", c.prefix)

	st := stack.New(certStackName, "us-east-1", c, nil)
	o, err := st.Run()
	if err != nil {
		return nil, err
	}
	return o, nil
}
