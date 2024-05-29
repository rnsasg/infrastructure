provider "aws" {
  region = "us-west-2"
}

resource "aws_iam_role" "eks_role" {
  name = "narayan_eks_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "eks.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      },
    ],
  })
}

# Create an EKS cluster 
resource "aws_eks_cluster" "narayan_eks_cluster" {
  name     = "narayan_eks_cluster"
  role_arn = aws_iam_role.eks_role.arn
  vpc_config {
    subnet_ids = ["", ""]
  }
}
