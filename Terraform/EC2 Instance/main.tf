# Configure AWS Provider 
provider "aws" {
  region = "us-west-2"
}

# Create an EC2 instance
resource "aws_intance" "example_ec2" {
  ami           = "ami-785db401"
  instance_type = "t2.micro"

  tags {
    Name = "terraform-ec2-example"
  }
}



