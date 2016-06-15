variable "aws_access_key_id" {
  description = "AWS Access Key ID"
  type        = "string"
}

variable "aws_secret_access_key" {
  description = "AWS Secret Access Key"
  type        = "string"
}

variable "aws_region" {
  description = "AWS Region"
  type        = "string"
}

variable "build_dir" {
  description = "Build directory"
  type        = "string"
  default     = "../build"
}
