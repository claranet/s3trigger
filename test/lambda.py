import boto3
import os

s3 = boto3.client('s3')
targetBucket = os.environ.get('TARGET_BUCKET')

"""Lambda function triggered from S3 which writes the keys to another bucket."""
def handle_creation(event, context):

    for r in event['Records']:
        key = r['s3']['object']['key']

        s3.put_object(Bucket=targetBucket, Key=key)
