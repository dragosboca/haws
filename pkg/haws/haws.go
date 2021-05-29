package haws

import (
	"fmt"
	"strings"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/route53"
)

type HAWS struct {
	oai          string
	bucketDomain string
	recordName   string
	certificate  string
	zoneId       string
	path         string
}

func New(oai string, bucketDomain string, certificateArn string, recordname string, zoneId string, path string) *HAWS {

	// format path for cloudformation
	p := fmt.Sprintf("/%s", strings.Trim(path, "/"))

	return &HAWS{
		oai:          oai,
		bucketDomain: bucketDomain,
		recordName:   recordname,
		certificate:  certificateArn,
		zoneId:       zoneId,
		path:         p,
	}
}

func (h *HAWS) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Resources["distribution"] = &cloudfront.Distribution{
		DistributionConfig: &cloudfront.Distribution_DistributionConfig{
			Aliases: []string{h.recordName},
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
					DomainName: h.bucketDomain,
					Id:         "cloudfront-hugo",
					OriginPath: h.path,
					S3OriginConfig: &cloudfront.Distribution_S3OriginConfig{
						OriginAccessIdentity: fmt.Sprintf("origin-access-identity/cloudfront/%s", h.oai),
					},
				},
			},
			ViewerCertificate: &cloudfront.Distribution_ViewerCertificate{
				AcmCertificateArn:      h.certificate,
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
		HostedZoneId: h.zoneId,
		Name:         h.recordName,
		Type:         "A",
	}

	t.Outputs["CloudFrontId"] = cloudformation.Output{
		Value:       cloudformation.Ref("distribution"),
		Description: "ARN of the cloudfront distribution",
	}

	return t
}
