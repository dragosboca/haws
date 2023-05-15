package haws

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/dragosboca/haws/pkg/stack"
)

type Haws struct {
	Prefix string
	Region string
	ZoneId string
	Domain string
	Record string
	Path   string

	dryRun bool

	Stacks map[string]*stack.Stack
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
	domain := *result.HostedZone.Name

	// trim trailing dot if any
	strings.TrimSuffix(domain, ".")

	return domain, nil
}

func New(prefix string, region string, record string, zoneId string, path string, dryRun bool) Haws {

	domain, err := getZoneDomain(zoneId)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	return Haws{
		Prefix: prefix,
		ZoneId: zoneId,
		Domain: domain,
		Region: region,
		Record: record,
		Path:   path,

		dryRun: dryRun,

		Stacks: make(map[string]*stack.Stack),
	}
}

func (h *Haws) DeployStack(name string, template stack.Template) error {
	h.Stacks[name] = stack.NewStack(template)
	if h.dryRun {
		fmt.Printf("DryRunning %s\n", name)
		return h.Stacks[name].DryRun()
	} else {
		fmt.Printf("Running %s\n", name)
		return h.Stacks[name].Run()
	}
}

func (h *Haws) GetStackOutput(name string, template stack.Template) error {
	h.Stacks[name] = stack.NewStack(template)
	return h.Stacks[name].GetOutputs()
}
