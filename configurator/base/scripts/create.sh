#!/bin/bash

terraform init
terraform workspace select $CLUSTER_NAME || terraform workspace new $CLUSTER_NAME
terraform apply -auto-approve -var worker_nodes=$SIZE_NODE_CLUSTER -var instance_type=$MACHINE_TYPE