# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY TWO EC2 INSTANCES THAT ALLOWS CONNECTIONS VIA SSH
# See test/terraform_ssh_example.go for how to write automated tests for this code.
# ---------------------------------------------------------------------------------------------------------------------

provider "aws" {
  region = "${var.aws_region}"
}

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY THE EC2 INSTANCE WITH A PUBLIC IP
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_instance" "example_public" {
  ami                    = "${var.ami_id}"
  instance_type          = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.example.id}"]
  key_name               = "${var.key_pair_name}"

  # This EC2 Instance has a public IP and will be accessible directly from the public Internet
  associate_public_ip_address = true

  tags {
    Name = "${var.instance_name}-public"
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY THE EC2 INSTANCE WITH A PRIVATE IP
# ---------------------------------------------------------------------------------------------------------------------


# ---------------------------------------------------------------------------------------------------------------------
# CREATE A SECURITY GROUP TO CONTROL WHAT REQUESTS CAN GO IN AND OUT OF THE EC2 INSTANCES
# ---------------------------------------------------------------------------------------------------------------------

resource "aws_security_group" "example" {
  name = "${var.instance_name}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = "${var.ssh_port}"
    to_port   = "${var.ssh_port}"
    protocol  = "tcp"

    # To keep this example simple, we allow incoming SSH requests from any IP. In real-world usage, you should only
    # allow SSH requests from trusted servers, such as a bastion host or VPN server.
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# LOOK UP THE LATEST UBUNTU AMI
# ---------------------------------------------------------------------------------------------------------------------
