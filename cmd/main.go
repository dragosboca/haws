package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dragosboca/haws/pkg/haws"
)

func main() {
	prefix := flag.String("prefix", "hugo", "A prefix that should be added to all resource names. It should be unique")
	region := flag.String("region", "eu-west-1", "AWS region for bucket and cloudfront distribution")
	record := flag.String("record", "", "Record name to be created in the dns zone. If it's ommited the apex of domain will point to cloudfront distribution. Record name should not include domain name")
	zoneId := flag.String("zone-id", "", "AWS Route53 zone ID for certificate validation and DNS record for cloudfront")
	path := flag.String("path-prefix", "", "Path prefix that will be appended by cloudfront to all requests")

	flag.Parse()

	h := haws.New(*prefix, *region, *record, *zoneId, *path)

	if err := h.AddStack("certificate", haws.NewCertificate(&h)); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	if err := h.AddStack("bucket", haws.NewBucket(&h)); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	if err := h.AddStack("cloudfront", haws.NewCdn(&h)); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	if err := h.AddStack("user", haws.NewIamUser(&h)); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	h.GenerateHugoConfig()
}
