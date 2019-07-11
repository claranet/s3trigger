#!/usr/bin/env bash
set -e

UNPROC=`jq -r '.modules[0].resources["aws_s3_bucket.unproc"].primary.id' terraform.tfstate`
PROC=`jq -r '.modules[0].resources["aws_s3_bucket.proc"].primary.id' terraform.tfstate`

../s3trigger -bucket $UNPROC

expected=25
actual=0
timeout=2

n=0
while [[ $n < 5 ]]
do

  actual=`aws s3 ls --recursive s3://$PROC/ | wc -l`
  if [[ $actual == $expected ]]; then
    break
  fi

  echo "Retrying in $timeout seconds."
  sleep $timeout
  timeout=$(( timeout * 2 ))
  n=$(( n + 1 ))
done

if [[ $actual == $expected ]]; then
    echo "PASSED: Found $actual files in bucket $PROC and expected $expected."
else
    echo "FAILED: Found $actual files in bucket $PROC but expected $expected."
    exit 1
fi
