# s3trigger [![Documentation](https://godoc.org/github.com/claranet/s3trigger?status.svg)](http://godoc.org/github.com/claranet/s3trigger)

This utiltiy simplifies re-invoking Lambda functions associated with a S3 bucket for all objects in the bucket. Rather
than pulling down each file and reuploading it or manipulating the object metadata, the Lambda event payload is
generated based on the bucket contents and send directly to the attached Lambda functions.

### Installation

If you have [installed Go](http://golang.org/doc/install.html), you can simply run this command
to install `s3trigger`:

```bash
go get github.com/claranet/s3trigger/cmd/s3trigger
```

You can also download the [latest](https://github.com/claranet/s3trigger/releases/latest) x64 release.

### Usage

AWS access is achieved using the default credential provider chain as part of the AWS SDK. As detailed in the
[Specifying Credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) section of the
SDK documentation, credentials are sought in environment variables, the shared credentials file and finally the instance
profile if you are running within AWS. Please note that you will need to specify your region, for example with the
`AWS_REGION` environment variable.

##### Example

Assuming you have [got your AWS access keys](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html),
you can export the three required environment variables and call `s3trigger`. `s3trigger` requires the `-bucket`
parameter with which you name the bucket you wish to trigger lambda events for.

```bash
export AWS_ACCESS_KEY_ID=*****
export AWS_SECRET_ACCESS_KEY=*****
export AWS_REGION=eu-west-1
s3trigger -bucket my-bucket 
```

##### Command line arguments

`s3trigger` expects the following command line argument:

- `-bucket` string **required**

  S3 bucket name to trigger Lambda functions for.

### Developing & Testing

Instead of using `go get`, you can clone this repository and use the `Makefile`. The following targets are available:

- `lint`

  Runs `golint` across the source reporting any style mistakes. If not already installed locally, you can run
  `go get -u golang.org/x/lint/golint` to install.
  
- `build`

  Runs `lint` and compiles the source to produce the `s3trigger` binary in the local directory.
  
- `test`

  Runs `build` then uses [Terraform](https://www.terraform.io/) to create two buckets, one with 25 objects (`unproc*`)
  and one with no objects (`proc*`). A lambda function is attached to the `unproc*` bucket which creates a key in the
  `proc*` bucket for each event it sees. The test calls `s3trigger` which should result in 25 objects being created with
  the same keys in the `proc*` bucket. Once the test has passed, the bucekts and lambda function are destroyed.
  Terraform is configured in the same way as `s3trigger` but requires additional IAM permissions as detailed below.

- `install` **default**

  Runs `test` and copies the local `s3trigger` to the user's `$GOPATH/bin` folder.
  
- `clean`

  Removes the local `s3trigger` if present and runs `terraform destroy` to ensure the buckets have been removed.

### IAM Permissions

The following IAM policy documents detail the minimum permissions required to execute `s3trigger` and `terraform`.

##### Minimum required permissions for `s3trigger`

```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": [
        "*"
      ],
      "Effect": "Allow"
    },
    {
      "Action": [
        "s3:GetBucketNotification",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::YOUR-BUCKET-NAME"
      ],
      "Effect": "Allow"
    }
  ]
}
```

##### Minimum required permissions for `terraform`

```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:AttachRolePolicy",
        "iam:CreateRole",
        "iam:CreatePolicy",
        "iam:DeletePolicy",
        "iam:DeleteRole",
        "iam:DetachRolePolicy",
        "iam:GetPolicy",
        "iam:GetPolicyVersion",
        "iam:GetRole",
        "iam:List*",
        "iam:PassRole",
        "lambda:AddPermission",
        "lambda:CreateFunction",
        "lambda:DeleteFunction",
        "lambda:GetFunction",
        "lambda:GetPolicy",
        "lambda:InvokeFunction",
        "lambda:ListVersionsByFunction",
        "lambda:RemovePermission",
        "s3:CreateBucket",
        "s3:DeleteBucket",
        "s3:DeleteObject",
        "s3:DeleteObjectVersion",
        "s3:Get*",
        "s3:ListBucket",
        "s3:ListBucketVersions",
        "s3:PutBucketNotification",
        "s3:PutBucketVersioning",
        "s3:PutObject"
      ],
      "Resource": [
        "*"
      ],
      "Effect": "Allow"
    }
  ]
}
``` 
