package haws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/dragosboca/haws/pkg/stack"
	"github.com/dragosboca/haws/pkg/template"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Part interface {
	stack.Runner
	template.Stack
}

type Haws struct {
	Prefix string
	Region string
	ZoneId string
	Domain string
	Record string
	Path   string

	dryRun bool

	bucket      Part
	certificate Part
	cloudfront  Part
	user        Part
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

func New(prefix string, region string, record string, zoneId string, path string, dryRun bool) *Haws {

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

	return &h
}

func (h *Haws) WithDefaults() *Haws {
	h.bucket = h.CreateBucket("bucket")
	h.certificate = h.CreateCertificate("certificate")
	h.cloudfront = h.CreateCdn("cloudfront")
	h.user = h.CreateIamUser("user")

	return h
}

func (h *Haws) Deploy() error {
	err := h.bucket.Deploy(h.dryRun, []*cloudformation.Parameter{})
	if err != nil {
		return err
	}

	err = h.certificate.Deploy(h.dryRun, []*cloudformation.Parameter{})
	if err != nil {
		return err
	}

	err = h.cloudfront.Deploy(h.dryRun, []*cloudformation.Parameter{{
		ParameterKey:   aws.String("CertificateArn"),
		ParameterValue: aws.String(h.certificate.OutputValue(h.certificate.GetExportName("Arn"))),
	}})
	if err != nil {
		return err
	}

	err = h.user.Deploy(h.dryRun, []*cloudformation.Parameter{})
	if err != nil {
		return err
	}

	return nil
}

func (h *Haws) GetOutputs() error {
	err := h.bucket.GetOutputs(h.dryRun)
	if err != nil {
		return err
	}

	err = h.certificate.GetOutputs(h.dryRun)
	if err != nil {
		return err
	}

	err = h.cloudfront.GetOutputs(h.dryRun)
	if err != nil {
		return err
	}

	err = h.user.GetOutputs(h.dryRun)
	if err != nil {
		return err
	}
	return nil
}
