provider "aws" {
  region = "us-west-2"
}

resource "aws_ecr_repository" "narayan_repo" {
  name                 = "hare_krishna_ecr_repo"
  image_tag_mutability = "MUTABLE"
}
