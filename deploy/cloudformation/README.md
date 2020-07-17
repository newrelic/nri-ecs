# Cloudformation

We have two stacks, one for registering the task and one for creating the
service. We couldn't use a single stack because then our customers wouldn't be
able to install the integration on multiple clusters, CloudFormation doesn't
allow to iterate over a list of values, meaning that customers launching the
stack couldn't only run the service in a single cluster. 

## Task registry stack

This stack lives in the `task/` directory. Is made up of a `master.yaml` file 
which imports two nested stacks:

1. `execution-role.yaml`: Handles the secrets and permissions. It creates:

    1. A secret with the license key given as parameter.
    1. A managed policy to access the license key. 
    1. An IAM role to be used as the ECS task execution role 

1. `task.yaml`: Registers the newrelic-infra task. It uses the execution role 
and the license key secret created by the `execution-role.yaml` stack.

To launch the stack click on the button.

[![button](https://dmhnzl5mp9mj6.cloudfront.net/application-management_awsblog/images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/create/review?templateURL=https://nr-downloads-main.s3.amazonaws.com/infrastructure_agent/integrations/ecs/cloudformation/task/master.yaml&stackName=NewRelicECSIntegration)

When we deploy a new version of the `nri-ecs` image we have to update the tag
in the `task.yaml` stack, then customers can do an update on the stack to 
register a new version of the `newrelic-infra` ECS task that uses the newly
deployed `nri-ecs` container.

## Service creation stack

This is the `service.yaml` stack, it creates the DAEMON service. You have to
specify the name of the cluster and the version to the `newrelic-infra` ECS
task. We couldn't linked both stacks by using cross-stack references because
then customers would have to delete all the service stacks to be able to update
the task stack.

[![button](https://dmhnzl5mp9mj6.cloudfront.net/application-management_awsblog/images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/create/review?templateURL=https://nr-downloads-main.s3.amazonaws.com/infrastructure_agent/integrations/ecs/cloudformation/service.yaml&NewRelicInfraTaskVersion=1)
