package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"log"
	"os"
	"s3trigger/pkg/s3trigger"
	"strings"
)

func main() {
	var bucket = flag.String("bucket", "", "Target S3 bucket (required)")
	var prefix = flag.String("prefix", "", "Only trigger for keys with matching prefix")

	flag.Parse()

	if strings.TrimSpace(*bucket) == "" {
		log.Print("bucket is a required parameter")
		flag.Usage()
		os.Exit(2)
	}

	if err := s3trigger.TriggerLambdasForBucketWithPrefix(*bucket, *prefix); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.Fatalf("triggering lambdas for bucket %s failed: %s", *bucket, awsErr.Message())
    	} else {
			log.Fatalf("triggering lambdas for bucket %s failed: %v", *bucket, err)
		}
	}
}
