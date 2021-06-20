package stack

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/goombaio/namegenerator"

	gf "github.com/awslabs/goformation/v4/cloudformation"
)

type Template interface {
	Build() *gf.Template
}

type Stack struct {
	Name           string
	Template       Template
	Region         string
	Parameters     []*cloudformation.Parameter
	CloudFormation *cloudformation.CloudFormation
}

type Output map[string]string

const EmptyChangeSet = "The submitted information didn't contain changes. Submit different information to create a change set."

func New(name string, region string, template Template, parameters []*cloudformation.Parameter) *Stack {
	var s *session.Session
	if region != "" {
		s = session.Must(session.NewSession(aws.NewConfig().WithRegion(region)))
	} else {
		s = session.Must(session.NewSession())
	}
	cf := cloudformation.New(s)

	return &Stack{
		Name:           name,
		Template:       template,
		Region:         region,
		Parameters:     parameters,
		CloudFormation: cf,
	}
}

func templateJson(t Template) (string, error) {
	template := t.Build()
	templateBody, err := template.JSON()
	if err != nil {
		fmt.Printf("Create template error: %s\n", err)
		return "", err
	}
	return string(templateBody), nil
}

// FIXME use type assertions on error
// FIXME FIXME: https://github.com/aws/aws-sdk/issues/44
func (st *Stack) stackExist() bool {
	_, err := st.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: &st.Name,
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

func (st *Stack) Deploy() error {

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	csName := nameGenerator.Generate()

	csType := "CREATE"
	if st.stackExist() {
		csType = "UPDATE"
		fmt.Printf("Updating stack: %s with changeset: %s\n", st.Name, csName)
	} else {
		fmt.Printf("Creating stack: %s with changeset: %s\n", st.Name, csName)
	}

	templateBody, err := templateJson(st.Template)
	if err != nil {
		fmt.Printf("Create template error: %s\n", err)
		return err
	}
	_, err = st.CloudFormation.CreateChangeSet(&cloudformation.CreateChangeSetInput{
		ClientToken:   &csName,
		ChangeSetName: &csName,
		ChangeSetType: &csType,
		Parameters:    st.Parameters,
		StackName:     &st.Name,
		TemplateBody:  aws.String(templateBody),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for the changeset %s creation to complete\n", csName)
	err = st.CloudFormation.WaitUntilChangeSetCreateComplete(&cloudformation.DescribeChangeSetInput{
		ChangeSetName: &csName,
		StackName:     &st.Name,
	})
	if err != nil {
		desc, err := st.CloudFormation.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: &csName,
			StackName:     &st.Name,
		})
		if err != nil {
			return err
		}
		if *desc.Status == cloudformation.ChangeSetStatusFailed && *desc.StatusReason == EmptyChangeSet {
			fmt.Printf("Deleting empty changeset %s\n", csName)
			_, err := st.CloudFormation.DeleteChangeSet(&cloudformation.DeleteChangeSetInput{
				ChangeSetName: &csName,
				StackName:     &st.Name,
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	fmt.Printf("Executing change set: %s on stack %s\n", csName, st.Name)
	_, err = st.CloudFormation.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      &csName,
		ClientRequestToken: &csName,
		StackName:          &st.Name,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for the changeset %s execution to complete\n", csName)
	if csType == "CREATE" {
		err = st.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: &st.Name,
		})
	} else {
		err = st.CloudFormation.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
			StackName: &st.Name,
		})
	}
	return err
}

func (st *Stack) Outputs() (Output, error) {
	stack, err := st.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: &st.Name})
	if err != nil {
		return nil, err
	}

	if len(stack.Stacks) > 1 {
		return nil, fmt.Errorf("multiple results for the same stack name %s", st.Name)
	}

	// Unpack outputs to a simpler structure
	ret := make(Output, 0)

	for _, a := range stack.Stacks[0].Outputs {
		ret[*a.OutputKey] = *a.OutputValue
	}

	return ret, nil
}

func (st *Stack) Run() (Output, error) {
	err := st.Deploy()
	if err != nil {
		return nil, err
	}
	o, err := st.Outputs()
	if err != nil {
		return nil, err
	}
	return o, nil
}
