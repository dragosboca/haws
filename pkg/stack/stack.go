package stack

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/tidwall/pretty"
)

type Stack struct {
	Template
	CloudFormation *cloudformation.CloudFormation
	Outputs        map[string]string
}

func NewStack(template Template) *Stack {
	return &Stack{
		Template: template,
		Outputs:  make(map[string]string),
	}
}

// Run creates or updates the stack
func (st *Stack) Run() error {
	var s *session.Session

	if st.GetRegion() != "" {
		s = session.Must(session.NewSession(aws.NewConfig().WithRegion(st.Template.GetRegion())))
	} else {
		s = session.Must(session.NewSession())
	}
	st.CloudFormation = cloudformation.New(s)

	templateBody, err := st.templateJson()
	if err != nil {
		return err
	}

	csName, csType, err := st.initialChangeSet(templateBody)
	if err != nil {
		return err
	}

	ok, err := st.waitForChangeSet(csName)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	if err := st.executeChangeSet(csName, csType); err != nil {
		return err
	}

	return st.GetOutputs()
}

// DryRun prints the template to stdout
func (st *Stack) DryRun() error {
	templateBody, err := st.templateJson()
	if err != nil {
		return err
	}

	for k, v := range st.GetDryRunOutputs() {
		st.Outputs[k] = v
	}

	fmt.Printf("%s\n", pretty.Color([]byte(templateBody), nil))

	return nil
}

// GetOutputs gets the outputs of the stack
func (st *Stack) GetOutputs() error {
	stack, err := st.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: st.GetStackName()})
	if err != nil {
		return err
	}

	if len(stack.Stacks) > 1 {
		return fmt.Errorf("multiple results for the same stack name %s", *st.GetStackName())
	}

	for _, a := range stack.Stacks[0].Outputs {
		st.Outputs[*a.OutputKey] = *a.OutputValue
	}

	return nil
}
