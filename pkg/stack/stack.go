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

type Stacker interface {
	Deploy(bool, string, []*cloudformation.Parameter) error
	GetOutputs(bool) error
	OutputValue(string) string
}

func newSession(region string) *session.Session {
	if region != "" {
		return session.Must(session.NewSession(aws.NewConfig().WithRegion(region)))
	}
	return session.Must(session.NewSession())
}

func NewStack(template Template) *Stack {
	return &Stack{
		Template:       template,
		CloudFormation: cloudformation.New(newSession(template.GetRegion())),
		Outputs:        make(map[string]string),
	}
}

func (st *Stack) run(params []*cloudformation.Parameter) error {
	templateBody, err := st.templateJson()
	if err != nil {
		return err
	}

	csName, csType, err := st.initialChangeSet(templateBody, params)
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

	return nil
}

func (st *Stack) dryRun() error {
	templateBody, err := st.templateJson()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", pretty.Color([]byte(templateBody), nil))

	return nil
}

func (st *Stack) GetOutputs(dryRun bool) error {
	if dryRun {
		for k, v := range st.GetDryRunOutputs() {
			st.Outputs[k] = v
		}
	} else {
		stack, err := st.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(st.GetStackName())})
		if err != nil {
			return err
		}

		if len(stack.Stacks) > 1 {
			return fmt.Errorf("multiple results for the same stack name %s", st.GetStackName())
		}

		for _, a := range stack.Stacks[0].Outputs {
			st.Outputs[*a.OutputKey] = *a.OutputValue
		}
	}

	return nil
}

func (st *Stack) Deploy(dryRun bool, name string, params []*cloudformation.Parameter) error {
	if dryRun {
		fmt.Printf("DryRunning %s\n", name)
		if err := st.dryRun(); err != nil {
			return err
		}
	} else {
		fmt.Printf("Running %s\n", name)
		if err := st.run(params); err != nil {
			return err
		}
	}
	return nil
}

func (st *Stack) OutputValue(name string) string {
	if val, ok := st.Outputs[name]; ok {
		return val
	}
	return ""
}
