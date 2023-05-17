package runner

import (
	"fmt"
	"github.com/dragosboca/haws/pkg/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/tidwall/pretty"
)

type ChangeSet struct {
	template.Stack
	CloudFormation *cloudformation.CloudFormation
	Outputs        map[string]string
}

type Runner interface {
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

func NewChangeSet(template template.Stack) *ChangeSet {
	return &ChangeSet{
		Stack:          template,
		CloudFormation: cloudformation.New(newSession(template.GetRegion())),
		Outputs:        make(map[string]string),
	}
}

func (cs *ChangeSet) run(params []*cloudformation.Parameter) error {
	templateBody, err := cs.templateJson()
	if err != nil {
		return err
	}

	csName, csType, err := cs.initialChangeSet(templateBody, params)
	if err != nil {
		return err
	}

	ok, err := cs.waitForChangeSet(csName)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	if err := cs.executeChangeSet(csName, csType); err != nil {
		return err
	}

	return nil
}

func (cs *ChangeSet) dryRun() error {
	templateBody, err := cs.templateJson()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", pretty.Color([]byte(templateBody), nil))

	return nil
}

func (cs *ChangeSet) GetOutputs(dryRun bool) error {
	if dryRun {
		for k, v := range cs.GetDryRunOutputs() {
			cs.Outputs[k] = v
		}
	} else {
		stack, err := cs.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(cs.GetStackName())})
		if err != nil {
			return err
		}

		if len(stack.Stacks) > 1 {
			return fmt.Errorf("multiple results for the same stack name %s", cs.GetStackName())
		}

		for _, a := range stack.Stacks[0].Outputs {
			cs.Outputs[*a.OutputKey] = *a.OutputValue
		}
	}

	return nil
}

func (cs *ChangeSet) Deploy(dryRun bool, name string, params []*cloudformation.Parameter) error {
	if dryRun {
		fmt.Printf("DryRunning %s\n", name)
		if err := cs.dryRun(); err != nil {
			return err
		}
	} else {
		fmt.Printf("Running %s\n", name)
		if err := cs.run(params); err != nil {
			return err
		}
	}
	return nil
}

func (cs *ChangeSet) OutputValue(name string) string {
	if val, ok := cs.Outputs[name]; ok {
		return val
	}
	return ""
}
