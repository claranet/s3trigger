//Package s3trigger provides utility functions to trigger S3 notifications on demand
package s3trigger

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
	"time"
)

var s3session = s3.New(session.New(), aws.NewConfig())
var lsession = lambda.New(session.New(), aws.NewConfig())

//TriggerLambdasForBucket invokes all lambda functions associated with
//bucket for each key contained in the bucket.
func TriggerLambdasForBucket (bucket string) error {
	return TriggerLambdasForBucketWithPrefix(bucket, "")
}

//TriggerLambdasForBucketWithPrefix invokes all lambda functions associated with
//bucket for each key starting with prefix in the bucket.
func TriggerLambdasForBucketWithPrefix (bucket string, prefix string) error {
	arns, err := GetLambdaArnsForBucket(bucket)
	if err != nil {
		return fmt.Errorf("get arns: %v", err)
	}

	return TriggerLambdaArnsForBucketWithPrefix(bucket, prefix, arns)
}

//GetLambdaArnsForBucket returns a list of lambda ARNs associated with bucket
func GetLambdaArnsForBucket (bucket string) ([]*string, error) {
	getNotif := &s3.GetBucketNotificationConfigurationRequest{
		Bucket: aws.String(bucket),
	}

	config, err := s3session.GetBucketNotificationConfiguration(getNotif)
	if err != nil {
		return nil, fmt.Errorf("s3 notification config: %v", err)
	}

	var arns []*string
	for _, conf := range config.LambdaFunctionConfigurations {
		arns = append(arns, conf.LambdaFunctionArn)
	}

	return arns, nil
}

//TriggerLambdaArnsForBucket invokes each lambda ARN for each key contained in bucket
//in batches of 10 (currently the maximum in AWS)
func TriggerLambdaArnsForBucket(bucket string, arns []*string) error {
	return TriggerLambdaArnsForBucketWithPrefix(bucket, "", arns)
}

//TriggerLambdaArnsForBucketWithPrefix invokes each lambda ARN for each key staring with the
//prefix and contained in bucket in batches of 10 (currently the maximum in AWS)
func TriggerLambdaArnsForBucketWithPrefix(bucket string, prefix string, arns []*string) error {
	listObjs := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	return triggerLambdaArnsForListObjectsInput(bucket, arns, listObjs)
}

func triggerLambdaArnsForListObjectsInput(bucket string, arns []*string, listObjs *s3.ListObjectsInput) error {

	errs := new(multierror.Error)

	err := s3session.ListObjectsPages(listObjs, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		var records []events.S3EventRecord
		for _, entry := range page.Contents {
			records = append(records, NewLambdaRecordForObject(bucket, entry))

			if len(records) == 10 {
				if err := InvokeLambdaArnsForRecords(records, arns); err != nil {
					errs = multierror.Append(errs, err)
				}
				records = records[:0]
			}
		}

		if len(records) > 0 {
			if err := InvokeLambdaArnsForRecords(records, arns); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		return aws.BoolValue(page.IsTruncated)
	})

	if err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

//NewLambdaRecordForObject generates the S3 ObjectCreated event record expected
//by a lambda function based on the bucket and object provided.
func NewLambdaRecordForObject (bucket string, object *s3.Object) events.S3EventRecord {
	bkt := events.S3Bucket{
		Name: bucket,
		Arn: "arn:aws:s3:::" + bucket,
	}

	obj := events.S3Object{
		Key: aws.StringValue(object.Key),
		Size: aws.Int64Value(object.Size),
		ETag: aws.StringValue(object.ETag),
	}

	entity := events.S3Entity{
		SchemaVersion: "1.0",
		Bucket: bkt,
		Object: obj,
	}

	record := events.S3EventRecord{
		EventVersion: "2.0",
		EventSource: "aws:s3",
		AWSRegion: aws.StringValue(s3session.Config.Region),
		EventTime: time.Now(),
		EventName: "ObjectCreated:CompleteMultipartUpload",
		S3: entity,
	}

	return record
}

//InvokeLambdaArnsForRecords calls InvokeLambdaArnForRecords for each of the arns
//with the event records
func InvokeLambdaArnsForRecords (records []events.S3EventRecord, arns []*string) error {
	errs := new(multierror.Error)

	for _, arn := range arns {
		if err := InvokeLambdaArnForRecords(records, arn); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

//InvokeLambdaArnForRecords calls the lambda ARN with the records as the S3 event
func InvokeLambdaArnForRecords (records []events.S3EventRecord, arn *string) error {

	event := events.S3Event{
		Records: records,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling records: %v", err)
	}

	invokeIn := &lambda.InvokeInput{
		FunctionName:   aws.String(*arn),
		InvocationType: aws.String("Event"),
		Payload:        payload,
	}

	if _, err := lsession.Invoke(invokeIn); err != nil {
		return fmt.Errorf("lambda invokation: %v", err)
	}

	return nil
}