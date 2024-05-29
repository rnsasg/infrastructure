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
