package haws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/dragosboca/haws/pkg/template"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/dragosboca/haws/pkg/runner"
)

type Haws struct {
	Prefix string
	Region string
	ZoneId string
	Domain string
	Record string
	Path   string

	dryRun    bool
	order     []string
	templates map[string]part // FIXME here
}

type params interface {
	setParametersValues(*Haws) []*cloudformation.Parameter
}

type part interface {
	template.Stack
	runner.Runner //FIXME rename
	params
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
func (h *Haws) addStack(name string, template part) {
	h.order = append(h.order, name)
	h.templates[name] = template
}

func New(prefix string, region string, record string, zoneId string, path string, dryRun bool) Haws {

	domain, err := getZoneDomain(zoneId)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	h := Haws{
		Prefix: prefix,
		ZoneId: zoneId,
		Domain: domain,
		Region: region,
		Record: record,
		Path:   path,
		dryRun: dryRun,
	}

	h.addStack("certificate", h.CreateCertificate()) // in US-EAST-1 => cannot be imported
	h.addStack("bucket", h.CreateBucket())           // in h.Region
	h.addStack("cloudfront", h.CreateCdn())          // in h.Region
	h.addStack("user", h.CreateIamUser())            // global
	return h
}

func (h *Haws) Deploy() error {
	for _, name := range h.order {
		if err := h.templates[name].Deploy(h.dryRun, name, h.templates[name].setParametersValues(h)); err != nil {
			return err
		}
		if err := h.templates[name].GetOutputs(h.dryRun); err != nil {
			return err
		}
	}
	return nil
}

func (h *Haws) RefreshOutputs() error {
	for _, name := range h.order {
		if err := h.templates[name].GetOutputs(h.dryRun); err != nil {
			return err
		}
	}
	return nil
}
