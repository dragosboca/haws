package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	region := flag.String("region", "eu-west-1", "AWS region for bucket and cloudfront distribution")
	zoneId := flag.String("zone-id", "", "AWS Route53 zone ID for certificate validation and DNS record for cloudfront")
	record := flag.String("record", "", "Record name to be created in the dns zone. If it's ommited the apex of domain will point to cloudfront distribution. Record name should not include domain name")
	prefix := flag.String("prefix", "hugo", "A prefix that should be added to all resource names. It should be unique")
	path := flag.String("path-prefix", "", "Path prefix that will be appended by cloudfront to all requests")

	flag.Parse()

	domain, err := getZoneDomain(*zoneId)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	rec := fmt.Sprintf("%s.%s", *record, domain)
	if *record == "" {
		rec = domain
	}

	certStackName := fmt.Sprintf("%sCertificate", *prefix)
	bucketStackName := fmt.Sprintf("%sBucket", *prefix)
	cloudfrontStackName := fmt.Sprintf("%s%sCloudfront", *prefix, *record)

	bucketName := fmt.Sprintf("haws-%s-%s-bucket", *prefix, strings.Replace(domain, ".", "-", -1))

	cOutput, err := deployCertificate(domain, zoneId, certStackName)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	s3Output, err := deployBucket(bucketName, bucketStackName, region)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	hOutput, reg, err := deployCloudFront(s3Output, cOutput, rec, zoneId, path, cloudfrontStackName, region)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	generateHugoConfig(path, bucketName, reg, hOutput)
}
