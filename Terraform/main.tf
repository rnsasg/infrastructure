terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.27"
    }
  }
  required_version = ">=0.14.9"
}

provider "aws" {
  alias  = "west2"
  region = "us-west-2"
}
