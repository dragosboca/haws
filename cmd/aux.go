package main

import (
	"fmt"
	"haws/pkg/bucket"
	"haws/pkg/certificate"
	"haws/pkg/haws"
	"haws/pkg/stack"
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

const hugoConfig = `
[[deployment.targets]]
    name = "haws"
    URL = "s3://{{ .BucketName }}?region={{ .Region }}&prefix={{ .OriginPath }}/"
    cloudFrontDistributionID = "{{ .CloudFrontId }}"
`

type deployment struct {
	BucketName   string
	Region       string
	CloudFrontId string
	OriginPath   string
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

func deployCertificate(domain string, zoneId *string, certStackName string) (stack.Output, error) {
	t := certificate.New(domain, *zoneId, []string{fmt.Sprintf("*.%s", domain)})
	st := stack.New(certStackName, "us-east-1", t, nil)
	err := st.Deploy()
	if err != nil {
		return nil, err
	}
	o, err := st.Outputs()
	if err != nil {
		return nil, err
	}
	return o, err
}

func deployBucket(bucketName string, bucketStackName string, region *string) (stack.Output, error) {
	t := bucket.New(bucketName)
	st := stack.New(bucketStackName, *region, t, nil)
	err := st.Deploy()
	if err != nil {
		return nil, err
	}
	o, err := st.Outputs()
	if err != nil {
		return nil, err
	}
	return o, err
}

func deployCloudFront(s3Output stack.Output, cOutput stack.Output, rec string, zoneId *string, path *string, cloudfrontStackName string, region *string) (stack.Output, string, error) {
	t := haws.New(s3Output["OAI"], s3Output["BucketDomain"], cOutput["CertificateArn"], rec, *zoneId, *path)
	st := stack.New(cloudfrontStackName, *region, t, nil)
	err := st.Deploy()
	if err != nil {
		return nil, "", err
	}
	o, err := st.Outputs()
	if err != nil {
		return nil, "", err
	}
	return o, *st.CloudFormation.Config.Region, err
}

func generateHugoConfig(path *string, bucketName string, reg string, hOutput stack.Output) {
	p := fmt.Sprintf("%s/", strings.Trim(*path, "/"))

	deploymentConfig := deployment{
		BucketName:   bucketName,
		Region:       reg,
		CloudFrontId: hOutput["CloudFrontId"],
		OriginPath:   p,
	}
	t := template.Must(template.New("deployment").Parse(hugoConfig))
	fmt.Printf("\n\n\n Use The folowing minimal configuration for hugo deployment\n")
	err := t.Execute(os.Stdout, deploymentConfig)
	if err != nil {
		fmt.Printf("Error executing template: %s", err)
		os.Exit(1)
	}
}
