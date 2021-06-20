package haws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"haws/pkg/cloudformation/bucket"
	"haws/pkg/cloudformation/cdn"
	"haws/pkg/cloudformation/certificate"
	"haws/pkg/cloudformation/user"
	"haws/pkg/stack"
	"os"
	"strings"
)

type Haws struct {
	prefix string

	certificate stack.Output
	bucket      stack.Output
	cloudfront  stack.Output
	user        stack.Output
	region      string

	bucketName string
	zoneId     string
	domain     string
	record     string
	path       string
}

type Deployable interface {
	Deploy() (stack.Output, error)
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
	if strings.HasSuffix(domain, ".") {
		domain = domain[:len(domain)-1]
	}

	return domain, nil
}

func New(prefix string, region string, record string, zoneId string, path string) Haws {

	domain, err := getZoneDomain(zoneId)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	return Haws{
		prefix: prefix,

		//iamUserName:         fmt.Sprintf("Haws%s%s", prefix, strings.Replace(domain, ".", "", -1)),

		zoneId: zoneId,
		domain: domain,
		region: region,
		record: record,
		path:   path,
	}
}

func (h *Haws) DeployCertificate() error {
	c := certificate.New(h.prefix, h.domain, h.zoneId, []string{fmt.Sprintf("*.%s", h.domain)})
	o, err := c.Deploy()
	if err != nil {
		return err
	}
	h.certificate = o
	return nil
}

func (h *Haws) DeployBucket() error {
	bucketName := fmt.Sprintf("Haws-%s-%s-bucket", h.prefix, strings.Replace(h.domain, ".", "-", -1))

	b := bucket.New(h.prefix, h.region, h.domain, bucketName)
	o, err := b.Deploy()
	if err != nil {
		return err
	}
	h.bucket = o
	return nil
}

func (h *Haws) DeployCloudFront() error {
	rec := fmt.Sprintf("%s.%s", h.record, h.domain)
	if h.record == "" {
		rec = h.domain
	}

	c := cdn.New(h.prefix, h.region, h.bucket["OAI"], h.bucket["BucketDomain"], h.certificate["CertificateArn"], rec, h.zoneId, h.path)
	o, err := c.Deploy()
	if err != nil {
		return err
	}
	h.cloudfront = o
	return nil
}

func (h *Haws) CreateIamUser() error {
	iamUserName := fmt.Sprintf("Haws%s%s", h.prefix, strings.Replace(h.domain, ".", "", -1))

	u := user.New(h.prefix, h.region, iamUserName, h.path, h.bucket["BucketArn"], h.cloudfront["CloudfrontArn"])
	o, err := u.Deploy()
	if err != nil {
		return err
	}
	h.user = o
	return nil
}
