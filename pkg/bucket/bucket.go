package bucket

import (
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/s3"
	"haws/pkg/customtags"
)

type Bucket struct {
	name string
}

func New(name string) *Bucket {
	return &Bucket{
		name: name,
	}
}

/// Policy Document
type document struct {
	Version   string
	Id        string
	Statement []statement
}

type principal map[string]string

type statement struct {
	Sid       string
	Effect    string
	Action    []string
	Principal principal
	Resource  []string
}

func newPolicy(id string) *document {
	return &document{
		Version:   "2008-10-17",
		Id:        id,
		Statement: []statement{},
	}
}

func (d *document) addStatement(sid string, s statement) {
	s.Sid = sid
	d.Statement = append(d.Statement, s)
}

func (b *Bucket) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Resources["oai"] = &cloudfront.CloudFrontOriginAccessIdentity{
		CloudFrontOriginAccessIdentityConfig: &cloudfront.CloudFrontOriginAccessIdentity_CloudFrontOriginAccessIdentityConfig{
			Comment: "haws oai",
		},
	}
	t.Outputs["OAI"] = cloudformation.Output{
		Value:       cloudformation.Ref("oai"),
		Description: "Origin Access Identity for Cloudfront",
	}

	t.Resources["bucket"] = &s3.Bucket{
		AccessControl:     "Private",
		BucketName:        b.name,
		CorsConfiguration: nil,
		Tags:              customtags.New(),
	}
	t.Outputs["BucketDomain"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "DomainName"),
		Description: "The domain name of the bucket",
	}

	doc := newPolicy("PolicyForCloudfrontPrivateContent")
	doc.addStatement("haws", statement{
		Effect: "Allow",
		Principal: principal{
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

	return t
}
