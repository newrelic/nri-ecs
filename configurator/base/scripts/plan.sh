#!/bin/bash

terraform init
terraform workspace select $CLUSTER_NAME || terraform workspace new $CLUSTER_NAME
terraform plan