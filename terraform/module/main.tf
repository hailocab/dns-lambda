provider "aws" {
  access_key = "${var.aws_access_key_id}"
  secret_key = "${var.aws_secret_access_key}"
  region     = "${var.aws_region}"
}

resource "aws_cloudwatch_event_target" "target" {
  rule = "${aws_cloudwatch_event_rule.rule.name}"
  arn  = "${aws_lambda_function.dns_lambda.arn}"
}

resource "aws_cloudwatch_event_rule" "rule" {
  name        = "dns_lambda"
  description = "DNS Lambda rule"

  event_pattern = <<PATTERN
{
  "source": [
    "aws.autoscaling"
  ],
  "detail-type": [
    "EC2 Instance Launch Successful",
    "EC2 Instance Terminate Successful"
  ]
}
PATTERN
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.dns_lambda.arn}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.rule.arn}"
}

resource "aws_iam_role" "role" {
  name = "dns_lambda_${var.aws_region}"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "role" {
  name = "dns_lambda_${var.aws_region}"
  role = "${aws_iam_role.role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*",
        "autoscaling:Describe*",
        "route53:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:route53:::*",
        "arn:aws:autoscaling:${var.aws_region}::*",
        "arn:aws:ec2:${var.aws_region}::*"
      ]
    }
  ]
}
EOF
}

resource "aws_lambda_function" "dns_lambda" {
  filename      = "${var.filename}"
  function_name = "dns_lambda"
  role          = "${aws_iam_role.role.arn}"
  handler       = "index.handle"
  runtime       = "nodejs4.3"
}
