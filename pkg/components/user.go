package components

import (
	"fmt"
	"strings"

	"github.com/dragosboca/haws/pkg/components/resources/customtags"
	"github.com/dragosboca/haws/pkg/components/resources/iampolicy"
	"github.com/dragosboca/haws/pkg/stack"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
)

type User struct {
	stack.TemplateComponent
	recordName string
	Path       string
	Prefix     string
}

type UserInput struct {
	Prefix        string
	Path          string
	Region        string
	Domain        string
	Record        string
	BucketName    string
	CloudfrontArn string
}

func NewIamUser(u *UserInput) *User {
	recordName := fmt.Sprintf("%s.%s", u.Record, u.Domain)
	if u.Record == "" {
		recordName = u.Domain
	}

	user := &User{
		Prefix:            u.Prefix,
		Path:              u.Path,
		TemplateComponent: stack.NewTemplate(u.Region),
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
				cloudformation.ImportValue(u.BucketName),
				cloudformation.Ref("Path"),
				"*",
			}),
			cloudformation.Join("/", []string{
				cloudformation.ImportValue(u.BucketName),
				cloudformation.Ref("Path"),
			}),
			cloudformation.ImportValue(u.CloudfrontArn),
		},
	})

	user.AddParameter("Path", cloudformation.Parameter{
		Type:        "String",
		Description: "The path in the bucket for the origin of the site",
	}, u.Path)

	user.AddParameter("Name", cloudformation.Parameter{
		Type:        "String",
		Description: "The name of the policy",
	}, fmt.Sprintf("Haws%s%s", u.Prefix, strings.ReplaceAll(u.Domain, ".", "")))

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
	}, "ACCESS_KEY")

	user.AddOutput("SecretKey", cloudformation.Output{
		Value:       cloudformation.GetAtt("accesskey", "SecretAccessKey"),
		Description: "SecretAccessKey for user",
	}, "SECRET_ACCESS_KEY")

	return user
}

func (u *User) GetExportName(output string) string {
	return fmt.Sprintf("HawsIamUser%s%s%s", output, strings.Title(u.Prefix), strings.Title(u.Path))
}

func (u *User) GetStackName() *string {
	stackName := fmt.Sprintf("%s-%s-iam-user", u.Prefix, u.recordName)
	return &stackName
}
