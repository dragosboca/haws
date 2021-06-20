package user

import (
	"fmt"
	"haws/pkg/customtags"
	"haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
)

type document struct {
	Version   string
	Id        string
	Statement []statement
}

type statement struct {
	Sid      string
	Effect   string
	Action   []string
	Resource []string
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

type User struct {
	name       string
	bucket     string
	cloudfront string
	path       string
	prefix     string
	region     string
}

func New(prefix string, region string, name string, path string, bucket string, cloudfront string) *User {
	return &User{
		name:       name,
		bucket:     bucket,
		cloudfront: cloudfront,
		path:       path,
		prefix:     prefix,
		region:     region,
	}
}

func (u *User) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	doc := newPolicy("PolicyForCloudfrontPrivateContent")
	doc.addStatement("haws", statement{
		Effect: "Allow",
		Action: []string{
			"s3:PutObject",
			"s3:PutBucketPolicy",
			"s3:ListBucket",
			"cloudfront:CreateInvalidation",
			"s3:GetBucketPolicy",
		},
		Resource: []string{
			fmt.Sprintf("%s/%s/*", u.bucket, u.path),
			fmt.Sprintf("%s/%s", u.bucket, u.path),
			u.cloudfront,
		},
	})

	t.Resources["user"] = &iam.User{
		Policies: []iam.User_Policy{
			{
				PolicyDocument: doc,
				PolicyName:     u.name,
			},
		},
		Tags:     customtags.New(),
		UserName: u.name,
	}

	t.Resources["accesskey"] = &iam.AccessKey{
		Serial:   0,
		UserName: cloudformation.Ref("user"),
	}

	t.Outputs["AccessKey"] = cloudformation.Output{
		Value:       cloudformation.Ref("accesskey"),
		Description: fmt.Sprintf("AccessKey for user %s", u.name),
	}
	t.Outputs["SecretKey"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("accesskey", "SecretAccessKey"),
		Description: fmt.Sprintf("SecretAccessKey for user %s", u.name),
	}

	return t
}

func (u *User) Deploy() (stack.Output, error) {

	userStackName := fmt.Sprintf("%sIamStackName", u.prefix)

	st := stack.New(userStackName, u.region, u, nil)
	o, err := st.Run()
	if err != nil {
		return nil, err
	}
	return o, nil
}
