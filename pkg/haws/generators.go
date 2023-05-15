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

func (h *Haws) getOutputByName(stack string, output string) string {
	return h.Stacks[stack].Outputs[h.Stacks[stack].GetExportName(output)]
}

func (h *Haws) GenerateHugoConfig() {
	deploymentConfig := deployment{
		BucketName:   h.getOutputByName("bucket", "Name"),
		Region:       h.Region,
		CloudFrontId: h.getOutputByName("cloudfront", "CloudFrontId"),
		OriginPath:   fmt.Sprintf("%s/", strings.Trim(h.Path, "/")),
	}
	t := template.Must(template.New("deployment").Parse(hugoConfig))
	fmt.Printf("\n\n\n Use The folowing minimal configuration for hugo deployment\n")
	err := t.Execute(os.Stdout, deploymentConfig)
	if err != nil {
		fmt.Printf("Error executing template: %s", err)
		os.Exit(1)
	}
}
