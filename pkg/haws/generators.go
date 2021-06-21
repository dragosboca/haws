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

//const hugoGithubFlow = `
//# This is a basic workflow to help you get started with Actions
//# Inspired by https://capgemini.github.io/development/Using-GitHub-Actions-and-Hugo-Deploy-to-Deploy-to-AWS/
//name: Build and Deploy Hugo Site
//
//# Controls when the action will run. Triggers the workflow on push
//# events but only for the master branch
//on:
//  push:
//    branches: [ master ]
//jobs:
//Build_and_Deploy:
//    runs-on: ubuntu-18.04
//    steps:
//    # Checks-out your repository under $GITHUB_WORKSPACE, so your job can access it
//    - uses: actions/checkout@v2
//    # Sets up Hugo
//    - name: Setup Hugo
//    uses: peaceiris/actions-hugo@v2
//    with:
//        hugo-version: '0.63.2'
//	# Builds repo
//	- name: Build
//	  run: hugo --minify
//	# Deploys built website to S3
//	- name: Deploy to S3
//	  run: hugo deploy --force --maxDeletes -1 --invalidateCDN
//	  env:
//		AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
//		AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
//`

type deployment struct {
	BucketName   string
	Region       string
	CloudFrontId string
	OriginPath   string
}

func (h *Haws) GetStackOutput(stack string, output string) string {
	return h.Stacks[stack].Outputs[h.Stacks[stack].GetOutputName(output)]
}

func (h *Haws) GenerateHugoConfig() {
	deploymentConfig := deployment{
		BucketName:   h.GetStackOutput("bucket", "Name"),
		Region:       h.Region,
		CloudFrontId: h.GetStackOutput("cloudfront", "CloudFrontId"),
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
