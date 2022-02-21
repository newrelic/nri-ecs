terraform {
  backend "s3" {
    bucket = "terraform-state-coreint-f33"
    key    = "terraform/demo4"
    region = "eu-central-1"
  }
}
