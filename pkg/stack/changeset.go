package stack

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/goombaio/namegenerator"
)

const EmptyChangeSet = "The submitted information didn't contain changes. Submit different information to create a change set."

func (cs *ChangeSet) stackExist() bool {
	_, err := cs.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(cs.GetStackName()),
	})
	return err == nil
}

func (cs *ChangeSet) templateJson() (string, error) {
	template := cs.Build()
	templateBody, err := template.JSON()
	if err != nil {
		fmt.Printf("Create template error: %s\n", err)
		return "", err
	}
	return string(templateBody), nil
}

func (cs *ChangeSet) initialChangeSet(templateBody string, params []*cloudformation.Parameter) (string, string, error) {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	csName := nameGenerator.Generate()

	csType := "CREATE"
	if cs.stackExist() {
		csType = "UPDATE"
		fmt.Printf("Updating stack: %s with changeset: %s\n", cs.GetStackName(), csName)
	} else {
		fmt.Printf("Creating stack: %s with changeset: %s\n", cs.GetStackName(), csName)
	}

	_, err := cs.CloudFormation.CreateChangeSet(&cloudformation.CreateChangeSetInput{
		ClientToken:   &csName,
		ChangeSetName: &csName,
		ChangeSetType: &csType,
		Parameters:    params, // FIXME this can be used to override parameters!!
		StackName:     aws.String(cs.GetStackName()),
		TemplateBody:  &templateBody,
	})
	if err != nil {
		return "", "", err
	}
	return csName, csType, nil
}

func (cs *ChangeSet) waitForChangeSet(csName string) (bool, error) {
	fmt.Printf("Waiting for the changeset %s creation to complete\n", csName)
	err := cs.CloudFormation.WaitUntilChangeSetCreateComplete(&cloudformation.DescribeChangeSetInput{
		ChangeSetName: &csName,
		StackName:     aws.String(cs.GetStackName()),
	})
	if err != nil {
		desc, err := cs.CloudFormation.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: &csName,
			StackName:     aws.String(cs.GetStackName()),
		})
		if err != nil {
			return false, err
		}
		if *desc.Status == cloudformation.ChangeSetStatusFailed && *desc.StatusReason == EmptyChangeSet {
			fmt.Printf("Deleting empty changeset %s\n", csName)
			_, err := cs.CloudFormation.DeleteChangeSet(&cloudformation.DeleteChangeSetInput{
				ChangeSetName: &csName,
				StackName:     aws.String(cs.GetStackName()),
			})
			if err != nil {
				return false, err
			}
			return true, nil
		} else {
			return false, err
		}
	}
	return false, nil
}

func (cs *ChangeSet) executeChangeSet(csName string, csType string) error {
	fmt.Printf("Executing change set: %s on stack %s\n", csName, cs.GetStackName())
	_, err := cs.CloudFormation.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      &csName,
		ClientRequestToken: &csName,
		StackName:          aws.String(cs.GetStackName()),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for the changeset %s execution to complete\n", csName)
	if csType == "CREATE" {
		err = cs.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(cs.GetStackName()),
		})
	} else {
		err = cs.CloudFormation.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(cs.GetStackName()),
		})
	}
	return err
}
