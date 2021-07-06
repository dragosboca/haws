package haws

import (
	"fmt"

	"strings"

	"github.com/dragosboca/haws/pkg/resources/bucketpolicy"
	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/stack"

	cfn "github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/s3"
)

type Bucket struct {
	*Haws
	stack.TemplateFactory
}

func NewBucket(h *Haws) *Bucket {
	return &Bucket{
		h,
		stack.NewTemplate(
			stack.WithParameter("BucketName", strings.ToLower(fmt.Sprintf("haws-%s-%s-bucket", h.Prefix, strings.Replace(h.Domain, ".", "-", -1))))),
	}
}

func (b *Bucket) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Parameters["BucketName"] = cloudformation.Parameter{
		Type:        "String",
		Description: "The name of the bucket with the content",
	}

	// OAI
	t.Resources["oai"] = &cloudfront.CloudFrontOriginAccessIdentity{
		CloudFrontOriginAccessIdentityConfig: &cloudfront.CloudFrontOriginAccessIdentity_CloudFrontOriginAccessIdentityConfig{
			Comment: cloudformation.Sub("haws oai for ${BucketName}"),
		},
	}

	// Bucket Itself
	t.Resources["bucket"] = &s3.Bucket{
		AccessControl:     "Private",
		BucketName:        cloudformation.Ref("BucketName"),
		CorsConfiguration: nil,
		Tags:              customtags.New(),
	}

	// Bucket Policy
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
	t.Resources["policy"] = &s3.BucketPolicy{
		Bucket:         cloudformation.Ref("bucket"),
		PolicyDocument: doc,
	}

	// Outputs
	t.Outputs["Domain"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "DomainName"),
		Description: "The domain name of the bucket",
		Export: &cloudformation.Export{
			Name: b.GetExportName("Domain"),
		},
	}

	t.Outputs["Arn"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "Arn"),
		Description: "The Arn of the bucket",
		Export: &cloudformation.Export{
			Name: b.GetExportName("Arn"),
		},
	}

	t.Outputs["Name"] = cloudformation.Output{
		Value:       cloudformation.Ref("BucketName"),
		Description: "The name of the bucket",
		Export: &cloudformation.Export{
			Name: b.GetExportName("Name"),
		},
	}

	t.Outputs["OAI"] = cloudformation.Output{
		Value:       cloudformation.Ref("oai"),
		Description: "Origin Access Identity for Cloudfront",
		Export: &cloudformation.Export{
			Name: b.GetExportName("Oai"),
		},
	}

	return t
}

func (b *Bucket) GetExportName(output string) string {
	return fmt.Sprintf("HawsBucket%s%s", output, strings.Title(b.Prefix))
}

func (b *Bucket) GetStackName() *string {
	stackName := fmt.Sprintf("%s-bucket", b.Prefix)
	return &stackName
}

func (b *Bucket) GetRegion() *string {
	return &b.Region
}

func (b *Bucket) GetParameters() []*cfn.Parameter {
	return b.Params
}

func (b *Bucket) DryRunOutputs() map[string]string {
	ret := make(map[string]string, 0)

	ret[b.GetExportName("Oai")] = "MockOai"
	ret[b.GetExportName("Domain")] = "mockdomain.com"
	ret[b.GetExportName("Arn")] = "aws:arn:s3:::mockBucket"
	ret[b.GetExportName("Name")] = "mockBucket"

	return ret
}
