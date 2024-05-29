
# Configure a AWS provider
provider "aws" {
  region = "us-west-2"
}

# Create an EC2 Instance
resource "aws_instance" "example_ec2" {
  ami                    = "ami-785db401"
  instance_type          = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.example_ec_sg.id}"]
  tags = {
    Name = "terraform-example-instance"
  }
  count     = 5
  user_data = <<-EOF
        #!/bin/bash
        echo "Hello world!" > index.html 
        nohup busybox httpd -f -p 8080 & 
        EOF
}

# Create a Security Group for an EC2 instance 

resource "aws_security_group" "example_ec_sg" {
  name = "narayan-example-security-group"
  ingress {
    to_port     = var.server_port
    from_port   = var.server_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

