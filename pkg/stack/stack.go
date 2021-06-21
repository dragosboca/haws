package stack

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type Output map[string]string

type Stack struct {
	Template
	CloudFormation *cloudformation.CloudFormation
	Outputs        Output
}

func NewStack(template Template) *Stack {
	var s *session.Session
	if *template.GetRegion() != "" {
		s = session.Must(session.NewSession(aws.NewConfig().WithRegion(*template.GetRegion())))
	} else {
		s = session.Must(session.NewSession())
	}
	cf := cloudformation.New(s)

	return &Stack{
		template,
		cf,
		make(Output, 0),
	}
}

// FIXME use type assertions on error
// FIXME FIXME: https://github.com/aws/aws-sdk/issues/44
func (st *Stack) stackExist() bool {
	_, err := st.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: st.Template.GetStackName(),
	})
	//if err != nil {
	//	if aerr, ok := err.(awserr.Error); ok{
	//		switch aerr.Code() {
	//		case cloudformation.AmazonCloudFormationException:
	//
	//		}
	//	}
	//}
	return err == nil
}

func (st *Stack) Run() error {
	templateBody, err := st.templateJson()
	if err != nil {
		return err
	}

	csName, csType, err := st.initialChangeSet(templateBody)
	if err != nil {
		return err
	}

	err = st.waitForChangeSet(csName, err)
	if err != nil {
		return err
	}

	err = st.executeChangeSet(csName, csType)
	if err != nil {
		return err
	}

	return st.getOutputs()
}
