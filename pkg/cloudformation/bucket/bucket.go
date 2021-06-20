package bucket

import (
	"fmt"
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v4/cloudformation/s3"
	"haws/pkg/customtags"
	"haws/pkg/stack"
)

type Bucket struct {
	name   string
	prefix string
	domain string
	region string
}

func New(prefix string, region string, domain string, name string) *Bucket {
	return &Bucket{
		name:   name,
		prefix: prefix,
		domain: domain,
		region: region,
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

	// OAI
	t.Resources["oai"] = &cloudfront.CloudFrontOriginAccessIdentity{
		CloudFrontOriginAccessIdentityConfig: &cloudfront.CloudFrontOriginAccessIdentity_CloudFrontOriginAccessIdentityConfig{
			Comment: fmt.Sprintf("haws oai for %s", b.name),
		},
	}
	t.Outputs["OAI"] = cloudformation.Output{
		Value:       cloudformation.Ref("oai"),
		Description: "Origin Access Identity for Cloudfront",
	}

	// Bucket Itself
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
	t.Outputs["BucketARN"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("bucket", "Arn"),
		Description: "The Arn of the bucket",
	}

	// Bucket Policy
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

func (b *Bucket) Deploy() (stack.Output, error) {

	bucketStackName := fmt.Sprintf("%sBucket", b.prefix)

	st := stack.New(bucketStackName, b.region, b, nil)
	o, err := st.Run()
	if err != nil {
		return nil, err
	}
	return o, nil
}
