# 
provider "aws" {
  region = "us-west-2"
}

resource "aws_s3_bucket" "narayan_bucket" {
  bucket = var.narayan_s3_bucket_name
  versioning {
    enabled = true
  }
  lifecycle {
    prevent_destroy = true
  }

}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket12312"
  acl    = "private"

  versioning {
    enabled = true
  }

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}
