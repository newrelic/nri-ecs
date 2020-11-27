# find out the image ID that AWS recommends for ECS in this region
data "aws_ssm_parameter" "amazon_linux_ecs" {
  name = "/aws/service/ecs/optimized-ami/amazon-linux-2/recommended/image_id"
}


# find the selected VPC
data "aws_vpc" "selected" {
  id = var.vpc_id
}

# find all subnets bound to the VPC
data "aws_subnet_ids" "public" {
  vpc_id = var.vpc_id

  tags = {
    owning_team = "coreint"
  }
}

# this user-data.sh script will make an EC2 node join the ECS cluster with the given name
data "template_file" "user_data" {
  template = file("${path.module}/templates/user-data.sh")

  vars = {
    cluster_name = local.cluster_name
  }
}

data "template_cloudinit_config" "config" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/x-shellscript"
    content      = data.template_file.user_data.rendered
  }
}
