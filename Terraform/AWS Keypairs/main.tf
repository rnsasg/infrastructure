resource "aws_instance" "web" {
  ami             = "ami-00ee4df451840fa9d" #Amazon Linux AMI
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.TF_SG.name]
  #first method
  key_name = "TF_key"


  tags = {
    Name = "Terraform Ec2"
  }
}

#keypair second method for Key_pair

resource "aws_key_pair" "TF_key" {
  key_name   = "TF_key"
  public_key = tls_private_key.rsa.public_key_openssh
}

resource "tls_private_key" "rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "TF-key" {
  content  = tls_private_key.rsa.private_key_pem
  filename = "tfkey"
}
