package haws

import (
	"fmt"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
	"haws/pkg/resources/customtags"
	"haws/pkg/resources/iampolicy"
	"haws/pkg/stack"
	"strings"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
)

type User struct {
	*Haws
	stack.TemplateFactory
	recordName string
}

func NewIamUser(h *Haws) *User {
	recordName := fmt.Sprintf("%s.%s", h.Record, h.Domain)
	if h.Record == "" {
		recordName = h.Domain
	}

	return &User{
		h,
		stack.NewTemplate(
			stack.WithParameter("Path", h.Path),
			stack.WithParameter("Name", fmt.Sprintf("Haws%s%s", h.Prefix, strings.Replace(h.Domain, ".", "", -1))),
		),
		recordName,
	}
}

func (u *User) Build() *cloudformation.Template {
	t := cloudformation.NewTemplate()

	doc := iampolicy.New("PolicyForCloudfrontPrivateContent")
	doc.AddStatement("haws", iampolicy.Statement{
		Effect: "Allow",
		Action: []string{
			"s3:PutObject",
			"s3:PutBucketPolicy",
			"s3:ListBucket",
			"cloudfront:CreateInvalidation",
			"s3:GetBucketPolicy",
		},
		Resource: []string{
			cloudformation.Join("/", []string{
				cloudformation.Ref(u.Stacks["bucket"].GetOutputName("Name")),
				cloudformation.Ref("Path"),
				"*",
			}),
			cloudformation.Join("/", []string{
				cloudformation.Ref(u.Stacks["bucket"].GetOutputName("Name")),
				cloudformation.Ref("Path"),
			}),
			cloudformation.Ref(u.Stacks["cloudfront"].GetOutputName("Arn")),
		},
	})

	t.Parameters["Path"] = cloudformation.Parameter{
		Type:        "String",
		Description: "The path in the bucket for the origin of the site",
	}
	t.Parameters["Name"] = cloudformation.Parameter{
		Type:        "String",
		Description: "The name of the policy",
	}

	t.Resources["user"] = &iam.User{
		Policies: []iam.User_Policy{
			{
				PolicyDocument: doc,
				PolicyName:     cloudformation.Ref("Name"),
			},
		},
		Tags:     customtags.New(),
		UserName: cloudformation.Ref("Name"),
	}

	t.Resources["accesskey"] = &iam.AccessKey{
		Serial:   0,
		UserName: cloudformation.Ref("user"),
	}

	t.Outputs["AccessKey"] = cloudformation.Output{
		Value:       cloudformation.Ref("accesskey"),
		Description: "AccessKey",
	}
	t.Outputs["SecretKey"] = cloudformation.Output{
		Value:       cloudformation.GetAtt("accesskey", "SecretAccessKey"),
		Description: "SecretAccessKey for user",
	}

	return t
}

func (u *User) GetOutputName(output string) string {
	return fmt.Sprintf("HawsIamUser%s%s%s", output, strings.Title(u.Prefix), strings.Title(u.Path))
}

func (u *User) GetStackName() *string {
	stackName := fmt.Sprintf("%s%sIamUser", u.Prefix, u.recordName)
	return &stackName
}

func (u *User) GetRegion() *string {
	return &u.Region
}

func (u *User) GetParameters() []*cfn.Parameter {
	return u.Params
}
