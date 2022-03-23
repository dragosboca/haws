package haws

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/resources/iampolicy"
	"github.com/dragosboca/haws/pkg/stack"

	cfn "github.com/aws/aws-sdk-go/service/cloudformation"

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
			stack.WithParameter("Name", fmt.Sprintf("Haws%s%s", h.Prefix, strings.ReplaceAll(h.Domain, ".", ""))),
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
				cloudformation.ImportValue(u.Stacks["bucket"].GetExportName("Name")),
				cloudformation.Ref("Path"),
				"*",
			}),
			cloudformation.Join("/", []string{
				cloudformation.ImportValue(u.Stacks["bucket"].GetExportName("Name")),
				cloudformation.Ref("Path"),
			}),
			cloudformation.ImportValue(u.Stacks["cloudfront"].GetExportName("Arn")),
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

func (u *User) GetExportName(output string) string {
	return fmt.Sprintf("HawsIamUser%s%s%s", output, strings.Title(u.Prefix), strings.Title(u.Path))
}

func (u *User) GetStackName() *string {
	stackName := fmt.Sprintf("%s-%s-iam-user", u.Prefix, u.recordName)
	return &stackName
}

func (u *User) GetRegion() *string {
	return &u.Region
}

func (u *User) GetParameters() []*cfn.Parameter {
	return u.Params
}

func (u *User) DryRunOutputs() map[string]string {
	ret := make(map[string]string)
	ret[u.GetExportName("AccessKey")] = "ACCESS_KEY"
	ret[u.GetExportName("SecretKey")] = "SECRET_ACCESS_KEY"
	return ret
}
