package haws

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/stack"

	cfn "github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/route53"
)

type Cdn struct {
	*Haws
	stack.TemplateFactory
	recordName string
}

func NewCdn(h *Haws) *Cdn {

	// format path for cloudformation
	p := fmt.Sprintf("/%s", strings.Trim(h.Path, "/"))

	// format rec
	recordName := fmt.Sprintf("%s.%s", h.Record, h.Domain)
	if h.Record == "" {
		recordName = h.Domain
	}

	return &Cdn{
		h,
		stack.NewTemplate(
			stack.WithParameter("RecordName", recordName),
			stack.WithParameter("CertificateArn", h.Stacks["certificate"].Outputs[h.Stacks["certificate"].GetExportName("Arn")]),
			stack.WithParameter("ZoneId", h.ZoneId),
			stack.WithParameter("Path", p),
		),
		recordName,
	}
}

//FIXME! Template

func (c *Cdn) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Parameters["RecordName"] = cloudformation.Parameter{
		Type:        "String",
		Description: "Record name for Route53 domain",
	}

	t.Parameters["CertificateArn"] = cloudformation.Parameter{
		Type:        "String",
		Description: "The ARN of the certificate generated in us-east-1 for cloudfront distribution",
	}

	t.Parameters["ZoneId"] = cloudformation.Parameter{
		Type:        "String",
		Description: "Route53 Zone Id",
	}

	t.Parameters["Path"] = cloudformation.Parameter{
		Type:        "String",
		Description: "The path in the bucket for the origin of the site",
	}

	t.Resources["distribution"] = &cloudfront.Distribution{
		DistributionConfig: &cloudfront.Distribution_DistributionConfig{
			Aliases: []string{
				cloudformation.Ref("RecordName"),
			},
			DefaultCacheBehavior: &cloudfront.Distribution_DefaultCacheBehavior{
				AllowedMethods: []string{"HEAD", "GET", "OPTIONS"},
				ForwardedValues: &cloudfront.Distribution_ForwardedValues{
					Cookies: &cloudfront.Distribution_Cookies{
						Forward: "none",
					},
				},
				MaxTTL:               86400,
				DefaultTTL:           3600,
				ViewerProtocolPolicy: "redirect-to-https",
				TargetOriginId:       "cloudfront-hugo",
			},
			Comment:           "Cloudfront for hugo website",
			DefaultRootObject: "index.html",
			Enabled:           true,
			HttpVersion:       "http2",
			IPV6Enabled:       true,
			Origins: []cloudfront.Distribution_Origin{
				{
					DomainName: cloudformation.ImportValue(c.Stacks["bucket"].GetExportName("Domain")),
					Id:         "cloudfront-hugo",
					OriginPath: c.Path,
					S3OriginConfig: &cloudfront.Distribution_S3OriginConfig{
						OriginAccessIdentity: cloudformation.Join("/", []string{
							"origin-access-identity/cloudfront",
							cloudformation.ImportValue(c.Stacks["bucket"].GetExportName("Oai")),
						}),
					},
				},
			},
			ViewerCertificate: &cloudfront.Distribution_ViewerCertificate{
				AcmCertificateArn:      cloudformation.Ref("CertificateArn"),
				MinimumProtocolVersion: "TLSv1.2_2019",
				SslSupportMethod:       "sni-only",
			},
		},
	}

	t.Resources["recordset"] = &route53.RecordSet{
		AliasTarget: &route53.RecordSet_AliasTarget{
			DNSName:      cloudformation.GetAtt("distribution", "DomainName"),
			HostedZoneId: "Z2FDTNDATAQYW2",
		},
		Comment:      "record for hugo website",
		HostedZoneId: cloudformation.Ref("ZoneId"),
		Name:         cloudformation.Ref("RecordName"),
		Type:         "A",
	}

	t.Outputs["CloudFrontId"] = cloudformation.Output{
		Value:       cloudformation.Ref("distribution"),
		Description: "ID cloudfront distribution",
		Export: &cloudformation.Export{
			Name: c.GetExportName("CloudFrontId"),
		},
	}

	t.Outputs["CloudFrontArn"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("distribution", "Arn"),
		Description: "ARN of the cloudfront distribution",
		Export: &cloudformation.Export{
			Name: c.GetExportName("CloudFrontArn"),
		},
	}

	return t
}

func (c *Cdn) GetExportName(output string) string {
	return fmt.Sprintf("HawsCloudfront%s%s%s", output, strings.Title(c.Prefix), strings.Title(c.Path))
}

func (c *Cdn) GetStackName() *string {

	stackName := fmt.Sprintf("%s-%s-cloudfront", c.Prefix, strings.Replace(c.recordName, ".", "-", -1))
	return &stackName
}

func (c *Cdn) GetRegion() *string {
	return &c.Region
}

func (c *Cdn) GetParameters() []*cfn.Parameter {
	return c.Params
}

func (c *Cdn) DryRunOutputs() map[string]string {
	ret := make(map[string]string)
	ret[c.GetExportName("CloudFrontId")] = "EDFDVBD632BHDS5"
	ret[c.GetExportName("CloudFrontArn")] = "arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5"
	return ret
}
