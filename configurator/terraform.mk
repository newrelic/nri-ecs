# Copyright 2019 New Relic Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
NATIVEOS	 := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH	 := $(shell go version | awk -F '[ /]' '{print $$5}')
TERRAFORM_DOCKER_IMAGE ?= hashicorp/terraform:0.13.5

CLUSTER_NAME ?= my-ecs-cluster
MACHINE_TYPE ?= t3.medium
SIZE_NODE_CLUSTER ?= 3

AWS_ACCESS_KEY_ID ?= "my-key"
AWS_SECRET_ACCESS_KEY ?= "my-secret"

all: plan

create-cluster:
	@echo "=== $(INTEGRATION) === [ create ]: Creating cluster..."
	docker run --rm -it --entrypoint= \
			-v $(CURDIR)/base:/terraform \
			-w /terraform \
			-e CLUSTER_NAME=$(CLUSTER_NAME) \
			-e SIZE_NODE_CLUSTER=$(SIZE_NODE_CLUSTER) \
			-e MACHINE_TYPE=$(MACHINE_TYPE) \
			-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
			-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
			$(TERRAFORM_DOCKER_IMAGE) sh scripts/create.sh


destroy-cluster:
	@echo "=== $(INTEGRATION) === [ destroy ]: Destroying cluster...."
	docker run --rm -it --entrypoint= \
			-v $(CURDIR)/configurator/base:/terraform \
			-w /terraform \
			-e CLUSTER_NAME=$(CLUSTER_NAME) \
			-e SIZE_NODE_CLUSTER=$(SIZE_NODE_CLUSTER) \
			-e MACHINE_TYPE=$(MACHINE_TYPE) \
			-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
			-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
			$(TERRAFORM_DOCKER_IMAGE) sh scripts/destroy.sh

plan:
	@echo "=== $(INTEGRATION) === [ plan ]: Showing plan...."
	docker run --rm -it --entrypoint= \
			-v $(CURDIR)/configurator/base:/terraform \
			-w /terraform \
			-e CLUSTER_NAME=$(CLUSTER_NAME) \
			-e SIZE_NODE_CLUSTER=$(SIZE_NODE_CLUSTER) \
			-e MACHINE_TYPE=$(MACHINE_TYPE) \
			-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
			-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
			$(TERRAFORM_DOCKER_IMAGE) sh scripts/plan.sh
				
.PHONY: all create-cluster destroy-cluster plan
