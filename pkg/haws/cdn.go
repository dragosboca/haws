package haws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	cloudformation2 "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/dragosboca/haws/pkg/template"
	"strings"

	"github.com/dragosboca/haws/pkg/runner"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/route53"
	"github.com/awslabs/goformation/v4/cloudformation/s3"
)

type Cdn struct {
	template.Template
	runner.ChangeSet
	recordName string
	Prefix     string
	Domain     string
	Path       string
}

func (h *Haws) CreateCdn() *Cdn {

	// format path for cloudformation
	path := fmt.Sprintf("/%s", strings.Trim(h.Path, "/"))

	// format rec
	recordName := fmt.Sprintf("%s.%s", h.Record, h.Domain)
	if h.Record == "" {
		recordName = h.Domain
	}

	cdn := &Cdn{
		Prefix:     h.Prefix,
		Domain:     h.Domain,
		Path:       h.Path,
		Template:   template.NewTemplate(h.Region),
		recordName: recordName,
	}

	cdn.AddParameter("RecordName", cloudformation.Parameter{
		Type:        "String",
		Description: "Record name for Route53 domain",
		Default:     recordName,
	})

	// FIXME populate parameters after initialization!!
	// FIXME maybe use parameter override?
	cdn.AddParameter("CertificateArn", cloudformation.Parameter{
		Type:        "String",
		Description: "The ARN of the certificate generated in us-east-1 for cloudfront distribution",
	})

	cdn.AddParameter("ZoneId", cloudformation.Parameter{
		Type:        "String",
		Description: "Route53 Zone Id",
		Default:     h.ZoneId,
	})

	cdn.AddParameter("Path", cloudformation.Parameter{
		Type:        "String",
		Description: "The path in the bucket for the origin of the site",
		Default:     path,
	})

	cdn.AddParameter("LogBucketName", cloudformation.Parameter{
		Type:        "String",
		Description: "The name of the bucket used for logs",
		Default:     fmt.Sprintf("%s-haws-logs-%s", cdn.Prefix, strings.ReplaceAll(cdn.Domain, ".", "-")),
	})

	cdn.AddResource("log_bucket", &s3.Bucket{
		BucketName:    cloudformation.Ref("LogBucketName"),
		AccessControl: "private",
	})

	cdn.AddResource("distribution", &cloudfront.Distribution{
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
			Logging: &cloudfront.Distribution_Logging{
				Bucket:         cloudformation.Ref("log_bucket"),
				Prefix:         cdn.Prefix,
				IncludeCookies: true,
			},
			Origins: []cloudfront.Distribution_Origin{
				{
					DomainName: cloudformation.ImportValue(h.templates["bucket"].GetExportName("Domain")),
					Id:         "cloudfront-hugo",
					OriginPath: cloudformation.Ref("Path"),
					S3OriginConfig: &cloudfront.Distribution_S3OriginConfig{
						OriginAccessIdentity: cloudformation.Join("/", []string{
							"origin-access-identity/cloudfront",
							cloudformation.ImportValue(h.templates["bucket"].GetExportName("Oai")),
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
	})

	cdn.AddResource("recordset", &route53.RecordSet{
		AliasTarget: &route53.RecordSet_AliasTarget{
			DNSName:      cloudformation.GetAtt("distribution", "DomainName"),
			HostedZoneId: "Z2FDTNDATAQYW2",
		},
		Comment:      "record for hugo website",
		HostedZoneId: cloudformation.Ref("ZoneId"),
		Name:         cloudformation.Ref("RecordName"),
		Type:         "A",
	})

	cdn.AddOutput("CloudFrontId", cloudformation.Output{
		Value:       cloudformation.Ref("distribution"),
		Description: "ID cloudfront distribution",
		Export: &cloudformation.Export{
			Name: cdn.GetExportName("CloudFrontId"),
		},
	}, "EDFDVBD632BHDS5")

	cdn.AddOutput("CloudFrontArn", cloudformation.Output{
		Value:       cloudformation.GetAtt("distribution", "Arn"),
		Description: "ARN of the cloudfront distribution",
		Export: &cloudformation.Export{
			Name: cdn.GetExportName("CloudFrontArn"),
		},
	}, "arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5")

	cdn.ChangeSet = *runner.NewChangeSet(cdn)
	return cdn
}

func (c *Cdn) GetExportName(output string) string {
	return fmt.Sprintf("HawsCloudfront%s%s%s", output, strings.Title(c.Prefix), strings.Title(c.Path))
}

func (c *Cdn) GetStackName() string {
	stackName := fmt.Sprintf("%s-%s-cloudfront", c.Prefix, strings.ReplaceAll(c.recordName, ".", "-"))
	return stackName
}

func (c *Cdn) setParametersValues(h *Haws) []*cloudformation2.Parameter {
	param := cloudformation2.Parameter{
		ParameterKey:   aws.String("CertificateArn"),
		ParameterValue: aws.String(h.templates["certificate"].OutputValue(h.templates["certificate"].GetExportName("Arn"))),
	}
	return []*cloudformation2.Parameter{&param}
}
