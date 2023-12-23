package haws

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

const hugoConfig = `
[[deployment.targets]]
    name = "Haws"
    URL = "s3://{{ .BucketName }}?region={{ .Region }}&prefix={{ .OriginPath }}/"
    cloudFrontDistributionID = "{{ .CloudFrontId }}"
`

type deployment struct {
	BucketName   string
	Region       string
	CloudFrontId string
	OriginPath   string
}

func (h *Haws) GenerateHugoConfig(region string, path string) {
	bucketName, err := h.GetOutputByName("bucket", "Name")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	cloudFrontId, err := h.GetOutputByName("cloudfront", "CloudFrontId")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	deploymentConfig := deployment{
		BucketName:   bucketName,
		Region:       region,
		CloudFrontId: cloudFrontId,
		OriginPath:   fmt.Sprintf("%s/", strings.Trim(path, "/")),
	}
	t := template.Must(template.New("deployment").Parse(hugoConfig))
	fmt.Printf("\n\n\n Use The folowing minimal configuration for hugo deployment\n")
	err = t.Execute(os.Stdout, deploymentConfig)
	if err != nil {
		fmt.Printf("Error executing template: %s", err)
		os.Exit(1)
	}
}
