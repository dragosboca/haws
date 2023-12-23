package haws

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/dragosboca/haws/pkg/components"
	"github.com/dragosboca/haws/pkg/stack"
)

type Haws struct {
	dryRun bool
	stacks map[string]*stack.Stack
}

func New(dryRun bool, prefix string, region string, zone_id string, bucketPath string, record string) Haws {
	domain, err := getZoneDomain(zone_id)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	h := Haws{
		dryRun: dryRun,
		stacks: make(map[string]*stack.Stack),
	}

	h.stacks["certificate"] = stack.NewStack(components.NewCertificate(&components.CertificateInput{
		Prefix: prefix,
		Region: region,
		Domain: domain,
		ZoneId: zone_id,
	}))

	h.stacks["bucket"] = stack.NewStack(components.NewBucket(&components.BucketInput{
		Prefix: prefix,
		Region: region,
		Domain: domain,
	}))

	h.stacks["cloudfront"] = stack.NewStack(components.NewCdn(&components.CdnInput{
		Prefix:         prefix,
		Path:           bucketPath,
		Region:         region,
		Domain:         domain,
		Record:         record,
		CertificateArn: h.stacks["certificate"].GetExportName("Arn"), // FIXME: this is not working because one cannot reference a resource from another region
		BucketDomain:   h.stacks["bucket"].GetExportName("Domain"),
		BucketOAI:      h.stacks["bucket"].GetExportName("Oai"),
		ZoneId:         zone_id,
	}))

	h.stacks["user"] = stack.NewStack(components.NewIamUser(&components.UserInput{
		Prefix:        prefix,
		Path:          bucketPath,
		Region:        region,
		Domain:        domain,
		Record:        record,
		BucketName:    h.stacks["bucket"].GetExportName("Name"),
		CloudfrontArn: h.stacks["cloudfront"].GetExportName("Arn"),
	}))
	return h
}

func (h *Haws) Deploy() error {
	stacks := []string{"certificate", "bucket", "cloudfront", "user"}
	for _, stack := range stacks {
		if stack == "cloudfronnt" { // ALL THIS STUPID THING BECAUSE CLOUDFORMATION DOES NOT SUPPORT CROSS REGION REFERENCES
			if err := h.GetStackOutput("certificate"); err != nil {
				return err
			}

			certificateArn, err := h.GetOutputByName("certificate", "certificateArn")
			if err != nil {
				return err
			}

			if err = h.SetStackParameterValue("certificate", "certificateArn", certificateArn); err != nil {
				return err
			}
		}
		if err := h.DeployStack(stack); err != nil {
			return err
		}

	}
	return nil
}

func (h *Haws) DeployStack(name string) error {
	if h.dryRun {
		fmt.Printf("DryRunning %s\n", name)
		return h.stacks[name].DryRun()
	} else {
		fmt.Printf("Running %s\n", name)
		return h.stacks[name].Run()
	}
}

func (h *Haws) GetStackOutput(name string) error {
	return h.stacks[name].GetOutputs()
}

func getZoneDomain(zoneId string) (string, error) {
	s := session.Must(session.NewSession())
	svc := route53.New(s)

	result, err := svc.GetHostedZone(&route53.GetHostedZoneInput{
		Id: &zoneId,
	})
	if err != nil {
		return "", err
	}
	// trim trailing dot if any
	domain := strings.TrimSuffix(*result.HostedZone.Name, ".")

	return domain, nil
}

func (h *Haws) SetStackParameterValue(stack string, parameter string, value string) error {
	if st, ok := h.stacks[stack]; ok {
		return st.SetParameterValue(parameter, value)
	}
	return fmt.Errorf("stack %s not found", stack)
}

func (h *Haws) GetOutputByName(stack string, output string) (string, error) {
	if st, ok := h.stacks[stack]; ok {
		return st.Outputs[st.GetExportName(output)], nil
	}
	return "", fmt.Errorf("stack %s not found", stack)

}
