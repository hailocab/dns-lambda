module "dns_lambda" {
  source = "../module"

  aws_access_key_id     = "${var.aws_access_key_id}"
  aws_secret_access_key = "${var.aws_secret_access_key}"
  aws_region            = "eu-west-1"

  filename = "../../build/archive.zip"
}
