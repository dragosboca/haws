package haws

import (
	"fmt"
	cloudformation2 "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/dragosboca/haws/pkg/template"
	"strings"

	"github.com/dragosboca/haws/pkg/resources/customtags"
	"github.com/dragosboca/haws/pkg/resources/iampolicy"
	"github.com/dragosboca/haws/pkg/runner"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
)

type User struct {
	template.Template
	runner.ChangeSet
	recordName string
	Path       string
	Prefix     string
}

func (h *Haws) CreateIamUser() *User {
	recordName := fmt.Sprintf("%s.%s", h.Record, h.Domain)
	if h.Record == "" {
		recordName = h.Domain
	}

	user := &User{
		Prefix:     h.Prefix,
		Path:       h.Path,
		Template:   template.NewTemplate(h.Region),
		recordName: recordName,
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
				cloudformation.ImportValue(h.templates["bucket"].GetExportName("Name")),
				cloudformation.Ref("Path"),
				"*",
			}),
			cloudformation.Join("/", []string{
				cloudformation.ImportValue(h.templates["bucket"].GetExportName("Name")),
				cloudformation.Ref("Path"),
			}),
			cloudformation.ImportValue(h.templates["cloudfront"].GetExportName("Arn")),
		},
	})

	user.AddParameter("Path", cloudformation.Parameter{
		Type:        "String",
		Description: "The path in the bucket for the origin of the site",
		Default:     h.Path})

	user.AddParameter("Name", cloudformation.Parameter{
		Type:        "String",
		Description: "The name of the policy",
		Default:     fmt.Sprintf("Haws%s%s", h.Prefix, strings.ReplaceAll(h.Domain, ".", ""))})

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

	user.ChangeSet = *runner.NewChangeSet(user)
	return user
}

func (u *User) GetExportName(output string) string {
	return fmt.Sprintf("HawsIamUser%s%s%s", output, strings.Title(u.Prefix), strings.Title(u.Path))
}

func (u *User) GetStackName() string {
	stackName := fmt.Sprintf("%s-%s-iam-user", u.Prefix, u.recordName)
	return stackName
}

func (u *User) setParametersValues(_ *Haws) []*cloudformation2.Parameter {
	return nil
}
