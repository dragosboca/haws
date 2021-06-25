package stack

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/tidwall/pretty"
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

func (st *Stack) DryRun() error {
	templateBody, err := st.templateJson()
	if err != nil {
		return err
	}

	for k, v := range st.DryRunOutputs() {
		st.Outputs[k] = v
	}

	fmt.Printf("%s\n", pretty.Color([]byte(templateBody), nil))

	return nil
}
