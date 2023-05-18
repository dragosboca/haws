package haws

import (
	"fmt"
	"github.com/dragosboca/haws/pkg/template"
	"strings"

	"github.com/dragosboca/haws/pkg/resources/bucketpolicy"
	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/s3"
)

type Bucket struct {
	template.Template
	*stack.ChangeSet
	Prefix string
}

func (h *Haws) CreateBucket(name string) *Bucket {
	bucket := &Bucket{
		Prefix:   h.Prefix,
		Template: template.NewTemplate(h.Region),
	}
	bucket.Name = name

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
			Default:     strings.ToLower(fmt.Sprintf("haws-%s-%s-bucket", h.Prefix, strings.ReplaceAll(h.Domain, ".", "-"))),
		},
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
	}, "mock.domain.com")

	bucket.AddOutput("Arn", cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "Arn"),
		Description: "The Arn of the bucket",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Arn"),
		},
	}, "aws:arn:s3:::mockBucket")

	bucket.AddOutput("Name", cloudformation.Output{
		Value:       cloudformation.Ref("BucketName"),
		Description: "The name of the bucket",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Name"),
		},
	}, "mockBucket")

	bucket.AddOutput("OAI", cloudformation.Output{
		Value:       cloudformation.Ref("oai"),
		Description: "Origin Access Identity for Cloudfront",
		Export: &cloudformation.Export{
			Name: bucket.GetExportName("Oai"),
		},
	}, "MockOai")

	bucket.ChangeSet = stack.NewChangeSet(bucket)
	return bucket
}

func (b *Bucket) GetExportName(output string) string {
	return fmt.Sprintf("HawsBucket%s%s", output, strings.Title(b.Prefix))
}

func (b *Bucket) GetStackName() string {
	stackName := fmt.Sprintf("%s-bucket", b.Prefix)
	return stackName
}
