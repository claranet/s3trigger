terraform {
  required_version = ">=0.11.0"
}

provider "aws" {
  version = "~> 2.0"
}

resource "aws_s3_bucket" "unproc" {
  bucket_prefix = "unproc"
  force_destroy = true
}

resource "aws_s3_bucket" "proc" {
  bucket_prefix = "proc"
  force_destroy = true
}

resource "aws_s3_bucket_object" "objs" {
  count   = 25
  bucket  = "${aws_s3_bucket.unproc.id}"
  key     = "somefile0${count.index}"
  content = "Test file ${count.index}"
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lambda_policy" {
  statement {
    actions = [
      "s3:*",
      "lambda:*",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role" "lambda_role" {
  name               = "lambdarole"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role.json}"
}

resource "aws_iam_policy" "lambda_policy" {
  policy = "${data.aws_iam_policy_document.lambda_policy.json}"
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = "${aws_iam_role.lambda_role.name}"
  policy_arn = "${aws_iam_policy.lambda_policy.arn}"
}

resource "aws_lambda_function" "lambda" {
  filename      = "lambda.zip"
  function_name = "s3_trigger_test"
  handler       = "lambda.handle_creation"
  role          = "${aws_iam_role.lambda_role.arn}"
  runtime       = "python3.6"

  environment {
    variables {
      TARGET_BUCKET = "${aws_s3_bucket.proc.id}"
    }
  }
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.lambda.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${aws_s3_bucket.unproc.arn}"
}

resource "aws_s3_bucket_notification" "bucket_lambda" {
  bucket = "${aws_s3_bucket.unproc.id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.lambda.arn}"
    events              = ["s3:ObjectCreated:*"]
  }
}
