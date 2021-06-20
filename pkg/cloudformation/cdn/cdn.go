package cdn

import (
	"fmt"
	"haws/pkg/stack"
	"strings"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/route53"
)

type Cdn struct {
	oai          string
	bucketDomain string
	recordName   string
	certificate  string
	zoneId       string
	path         string
	prefix       string
	region       string
}

func New(prefix string, region string, oai string, bucketDomain string, certificateArn string, recordname string, zoneId string, path string) *Cdn {

	// format path for cloudformation
	p := fmt.Sprintf("/%s", strings.Trim(path, "/"))

	return &Cdn{
		oai:          oai,
		bucketDomain: bucketDomain,
		recordName:   recordname,
		certificate:  certificateArn,
		zoneId:       zoneId,
		path:         p,
		prefix:       prefix,
		region:       region,
	}
}

func (c *Cdn) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Resources["distribution"] = &cloudfront.Distribution{
		DistributionConfig: &cloudfront.Distribution_DistributionConfig{
			Aliases: []string{c.recordName},
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
					DomainName: c.bucketDomain,
					Id:         "cloudfront-hugo",
					OriginPath: c.path,
					S3OriginConfig: &cloudfront.Distribution_S3OriginConfig{
						OriginAccessIdentity: fmt.Sprintf("origin-access-identity/cloudfront/%s", c.oai),
					},
				},
			},
			ViewerCertificate: &cloudfront.Distribution_ViewerCertificate{
				AcmCertificateArn:      c.certificate,
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
		HostedZoneId: c.zoneId,
		Name:         c.recordName,
		Type:         "A",
	}

	t.Outputs["CloudFrontId"] = cloudformation.Output{
		Value:       cloudformation.Ref("distribution"),
		Description: "ID cloudfront distribution",
	}
	t.Outputs["CloudFrontArn"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("distribution", "Arn"),
		Description: "ARN of the cloudfront distribution",
	}

	return t
}

func (c *Cdn) Deploy() (stack.Output, error) {

	cloudfrontStackName := fmt.Sprintf("%s%sCloudfront", c.prefix, c.recordName)

	st := stack.New(cloudfrontStackName, c.region, c, nil)
	o, err := st.Run()
	if err != nil {
		return nil, err
	}
	return o, nil
}
