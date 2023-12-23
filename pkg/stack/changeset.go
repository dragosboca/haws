package stack

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/goombaio/namegenerator"
)

const EmptyChangeSet = "The submitted information didn't contain changes. Submit different information to create a change set."

// FIXME use type assertions on error
// FIXME FIXME: https://github.com/aws/aws-sdk/issues/44
func (st *Stack) stackExist() bool {
	_, err := st.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: st.GetStackName(),
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

// templateJson returns the template as a JSON string
// return: string - the template as a JSON string
// return: error - the error if any
func (st *Stack) templateJson() (string, error) {
	template := st.Build()
	templateBody, err := template.JSON()
	if err != nil {
		fmt.Printf("Create template error: %s\n", err)
		return "", err
	}
	return string(templateBody), nil
}

// initialChangeSet creates the initial changeset
// param: templateBody - the template body
// return: csName - the name of the changeset
// return: csType - the type of the changeset (CREATE or UPDATE)
// return: error - the error if any
func (st *Stack) initialChangeSet(templateBody string) (string, string, error) {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	csName := nameGenerator.Generate()

	csType := "CREATE"
	if st.stackExist() {
		csType = "UPDATE"
		fmt.Printf("Updating stack: %s with changeset: %s\n", *st.GetStackName(), csName)
	} else {
		fmt.Printf("Creating stack: %s with changeset: %s\n", *st.GetStackName(), csName)
	}

	_, err := st.CloudFormation.CreateChangeSet(&cloudformation.CreateChangeSetInput{
		ClientToken:   &csName,
		ChangeSetName: &csName,
		ChangeSetType: &csType,
		Parameters:    st.GetParameters(),
		StackName:     st.GetStackName(),
		TemplateBody:  &templateBody,
	})
	if err != nil {
		return "", "", err
	}
	return csName, csType, nil
}

// waitForChangeSet returns true if the changeset is empty and should be deleted
// param: csName - the name of the changeset
// return: bool - true if the changeset is empty and should be deleted
// return: error - the error if any
func (st *Stack) waitForChangeSet(csName string) (bool, error) {
	fmt.Printf("Waiting for the changeset %s creation to complete\n", csName)
	err := st.CloudFormation.WaitUntilChangeSetCreateComplete(&cloudformation.DescribeChangeSetInput{
		ChangeSetName: &csName,
		StackName:     st.GetStackName(),
	})
	if err != nil {
		desc, err := st.CloudFormation.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: &csName,
			StackName:     st.GetStackName(),
		})
		if err != nil {
			return false, err
		}
		if *desc.Status == cloudformation.ChangeSetStatusFailed && *desc.StatusReason == EmptyChangeSet {
			fmt.Printf("Deleting empty changeset %s\n", csName)
			_, err := st.CloudFormation.DeleteChangeSet(&cloudformation.DeleteChangeSetInput{
				ChangeSetName: &csName,
				StackName:     st.GetStackName(),
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

// executeChangeSet executes the changeset
// param: csName - the name of the changeset
// param: csType - the type of the changeset (CREATE or UPDATE)
// return: error - the error if any
func (st *Stack) executeChangeSet(csName string, csType string) error {
	fmt.Printf("Executing change set: %s on stack %s\n", csName, *st.GetStackName())
	_, err := st.CloudFormation.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      &csName,
		ClientRequestToken: &csName,
		StackName:          st.GetStackName(),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for the changeset %s execution to complete\n", csName)
	if csType == "CREATE" {
		err = st.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: st.GetStackName(),
		})
	} else {
		err = st.CloudFormation.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
			StackName: st.GetStackName(),
		})
	}
	return err
}
