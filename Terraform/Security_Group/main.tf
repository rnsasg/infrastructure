provider "aws" {
  region = "us-west-2"
}

resource "aws_security_group" "allowweb" {
  name        = "allow_web_traffic"
  description = "allow web inbound traffic"
  vpc_id      = "vpc-12345678"

  # Define Ingress rule 
  ingress {
    description = "HTTPS from VPC"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_web"
  }
}
