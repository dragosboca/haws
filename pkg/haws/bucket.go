package haws

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/resources/bucketpolicy"
	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/s3"
)

type Bucket struct {
	*Haws
	stack.TemplateComponent
}

func NewBucket(h *Haws) *Bucket {
	bucket := &Bucket{
		Haws:              h,
		TemplateComponent: stack.NewTemplate(),
	}

	doc := bucketpolicy.New("PolicyForCloudfrontPrivateContent")
	doc.AddStatement("haws", bucketpolicy.Statement{
		Effect: "Allow",
		Principal: bucketpolicy.Principal{
			"AWS": cloudformation.Sub("arn:aws:iam::cloudfront:user/CloudFront Origin Access Identity ${oai}"),
		},
		Action: []string{"s3:GetObject"},
		Resource: []string{
			cloudformation.Join("/", []string{cloudformation.GetAtt("bucket", "Arn"), "*"}),
			cloudformation.GetAtt("bucket", "Arn"),
		},
	})

	bucket.AddParameter("BucketName",
		cloudformation.Parameter{
			Type:        "String",
			Description: "The name of the bucket with the content",
		},
		strings.ToLower(fmt.Sprintf("haws-%s-%s-bucket", h.Prefix, strings.ReplaceAll(h.Domain, ".", "-"))),
	)

	bucket.AddResource("oai", &cloudfront.CloudFrontOriginAccessIdentity{
		CloudFrontOriginAccessIdentityConfig: &cloudfront.CloudFrontOriginAccessIdentity_CloudFrontOriginAccessIdentityConfig{
			Comment: cloudformation.Sub("haws oai for ${BucketName}"),
		}})

	bucket.AddResource("bucket", &s3.Bucket{
		AccessControl:     "Private",
		BucketName:        cloudformation.Ref("BucketName"),
		CorsConfiguration: nil,
		Tags:              customtags.New(),
	})

	bucket.AddResource("policy", &s3.BucketPolicy{
		Bucket:         cloudformation.Ref("bucket"),
		PolicyDocument: doc,
	})

	bucket.AddOutput("Domain", cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "DomainName"),
		Description: "The domain name of the bucket",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Domain"),
		},
	})

	bucket.AddOutput("Arn", cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "Arn"),
		Description: "The Arn of the bucket",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Arn"),
		},
	})

	bucket.AddOutput("Name", cloudformation.Output{
		Value:       cloudformation.Ref("BucketName"),
		Description: "The name of the bucket",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Name"),
		},
	})

	bucket.AddOutput("OAI", cloudformation.Output{
		Value:       cloudformation.Ref("oai"),
		Description: "Origin Access Identity for Cloudfront",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Oai"),
		},
	})

	return bucket
}

func (b *Bucket) GetExportName(output string) string {
	return fmt.Sprintf("HawsBucket%s%s", output, strings.Title(b.Prefix))
}

func (b *Bucket) GetStackName() *string {
	stackName := fmt.Sprintf("%s-bucket", b.Prefix)
	return &stackName
}

func (b *Bucket) DryRunOutputs() map[string]string {
	ret := make(map[string]string)

	ret[b.GetExportName("Oai")] = "MockOai"
	ret[b.GetExportName("Domain")] = "mock.domain.com"
	ret[b.GetExportName("Arn")] = "aws:arn:s3:::mockBucket"
	ret[b.GetExportName("Name")] = "mockBucket"

	return ret
}
