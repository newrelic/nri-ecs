variable "AWS_REGION" {
  default = "eu-central-1"
}

variable "ECS_AMIS" {
  type = map(string)
  default = {
    eu-central-1 = "ami-065c1e34da68f2b02"
  }
}

variable "ECS_INSTANCE_TYPE" {
  default = "t2.micro"
}

variable "SSH_KEY_NAME" {
  description = "Which SSH key to use"
  default     = "coreint-ecs"
}