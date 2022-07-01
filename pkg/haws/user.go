package haws

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/resources/iampolicy"
	"github.com/dragosboca/haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
)

type User struct {
	*Haws
	stack.TemplateComponent
	recordName string
}

func NewIamUser(h *Haws) *User {
	recordName := fmt.Sprintf("%s.%s", h.Record, h.Domain)
	if h.Record == "" {
		recordName = h.Domain
	}

	user := &User{
		Haws:              h,
		TemplateComponent: stack.NewTemplate(),
		recordName:        recordName,
	}

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
				cloudformation.ImportValue(h.Stacks["bucket"].GetExportName("Name")),
				cloudformation.Ref("Path"),
				"*",
			}),
			cloudformation.Join("/", []string{
				cloudformation.ImportValue(h.Stacks["bucket"].GetExportName("Name")),
				cloudformation.Ref("Path"),
			}),
			cloudformation.ImportValue(h.Stacks["cloudfront"].GetExportName("Arn")),
		},
	})

	user.AddParameter("Path", cloudformation.Parameter{
		Type:        "String",
		Description: "The path in the bucket for the origin of the site",
	}, h.Path)

	user.AddParameter("Name", cloudformation.Parameter{
		Type:        "String",
		Description: "The name of the policy",
	}, fmt.Sprintf("Haws%s%s", h.Prefix, strings.ReplaceAll(h.Domain, ".", "")))

	user.AddResource("user", &iam.User{
		Policies: []iam.User_Policy{
			{
				PolicyDocument: doc,
				PolicyName:     cloudformation.Ref("Name"),
			},
		},
		Tags:     customtags.New(),
		UserName: cloudformation.Ref("Name"),
	})

	user.AddResource("accesskey", &iam.AccessKey{
		Serial:   0,
		UserName: cloudformation.Ref("user"),
	})

	user.AddOutput("AccessKey", cloudformation.Output{
		Value:       cloudformation.Ref("accesskey"),
		Description: "AccessKey",
	})

	user.AddOutput("SecretKey", cloudformation.Output{
		Value:       cloudformation.GetAtt("accesskey", "SecretAccessKey"),
		Description: "SecretAccessKey for user",
	})

	return user
}

func (u *User) GetExportName(output string) string {
	return fmt.Sprintf("HawsIamUser%s%s%s", output, strings.Title(u.Prefix), strings.Title(u.Path))
}

func (u *User) GetStackName() *string {
	stackName := fmt.Sprintf("%s-%s-iam-user", u.Prefix, u.recordName)
	return &stackName
}

func (u *User) DryRunOutputs() map[string]string {
	ret := make(map[string]string)
	ret[u.GetExportName("AccessKey")] = "ACCESS_KEY"
	ret[u.GetExportName("SecretKey")] = "SECRET_ACCESS_KEY"
	return ret
}
