provider "aws" {
  profile = "801306408012/okta_coreintegrationsteam"
  region  = "us-east-1"
  version = "~> 3.0"
}

provider "template" {
  version = "~> 2.1"
}

provider "null" {
  version = "~> 2.1"
}

provider "random" {
  version = "~> 2.2"
}

locals {
  cluster_name = terraform.workspace

  environment = "dev"
  team        = "coreint"
  product     = "infrastructure"

  # This is the convention we use to know what belongs to each other
  ec2_resources_name = "${local.cluster_name}-${local.environment}"
  common_tags = {
    environment = local.environment
    owning_team = local.team
    product     = local.product
  }
}

# Create the ECS cluster
module "ecs" {
  source  = "terraform-aws-modules/ecs/aws"
  version = "~> 5.0"

  name = local.cluster_name
  tags = local.common_tags
}

terraform {
  backend "s3" {
    bucket = "coreint-ecs-state-store"
    key    = "terraform.tfstate"
    region = "us-east-1"
  }
}

module "this" {
  source  = "terraform-aws-modules/autoscaling/aws"
  version = "~> 3.4"

  name = local.ec2_resources_name

  # Launch configuration
  lc_name = local.ec2_resources_name

  image_id             = data.aws_ssm_parameter.amazon_linux_ecs.value
  instance_type        = var.instance_type
  security_groups      = [aws_security_group.cluster_sg.id]
  iam_instance_profile = aws_iam_instance_profile.ssm_profile.id

  user_data = data.template_cloudinit_config.config.rendered
  # key_name  = var.ssh_key_name

  # Auto scaling group
  asg_name                  = local.ec2_resources_name
  vpc_zone_identifier       = data.aws_subnet_ids.public.ids
  health_check_type         = "EC2"
  min_size                  = 0
  max_size                  = var.worker_nodes 
  desired_capacity          = var.worker_nodes
  wait_for_capacity_timeout = 0

  root_block_device = [{
    volume_size           = 30
    volume_type           = "gp2"
    delete_on_termination = true
  }]

  tags = [
    {
      key                 = "environment"
      value               = local.environment
      propagate_at_launch = true
    },
    {
      key                 = "owning_team"
      value               = local.team
      propagate_at_launch = true
    },
    {
      key                 = "product"
      value               = local.product
      propagate_at_launch = true
    },
    {
      key                 = "ECS_cluster"
      value               = local.cluster_name
      propagate_at_launch = true
    },
  ]
}

resource "aws_security_group" "cluster_sg" {
  name        = "ecs_${local.cluster_name}_allow_ssh"
  description = "Allow SSH traffic from NR offices in BCN & PDX"
  vpc_id      = data.aws_vpc.selected.id

  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
    cidr_blocks = [
      "213.27.240.242/32", # Office - Barcelona (Glories)
      "38.104.105.178/32", # Office - PDX 1
      "4.15.128.122/32",   # Office - PDX 2
    ]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = local.common_tags
}

resource "aws_iam_role" "ssm_role" {
  name = "${local.cluster_name}_ecs_instance_role"
  path = "/ecs/"

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": ["ec2.amazonaws.com"]
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "ssm_profile" {
  name = "${local.cluster_name}_ecs_instance_profile"
  role = aws_iam_role.ssm_role.name
}

resource "aws_iam_role_policy_attachment" "ecs_ecs" {
  role       = aws_iam_role.ssm_role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_role_policy_attachment" "ecs_ec2_cloudwatch_role" {
  role       = aws_iam_role.ssm_role.id
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

resource "aws_iam_role_policy_attachment" "ecs_ec2_ssm_role" {
  role       = aws_iam_role.ssm_role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM"
}