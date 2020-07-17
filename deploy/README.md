# Deployments

## Configuring the license key

To properly use the task definitions below, the license key has to be configured.
We recommend the use of the AWS System Manager's Parameter Store to securely
store your license key. Look for the placeholder
`<SYSTEM_MANAGER_LICENSE_PARAMETER_NAME>` in the example JSON and replace it
with the name of the parameter in System Manager.

Note that the ECS Task Execution Role needs to have permissions to **read**
resources from the System Manager, otherwise it won't be able to fetch the
secret's value and the containers will not start. **We recommend that you select
the task execution role via the builder interface.**

It's possible to configure the license key as an environment variable but we
don't recommend this setup.

## Task definition: Infra Agent on ECS EC2 as a Daemon Service

In `task_definition_example.json` there is an example of an ECS task definition
that can be used to deploy the Infrastructure Agent as a Daemon on ECS EC2 instances.

### Compatibility

* It should be deployed with a **DAEMON** Service using with **EC2 Launch Type**.
* Currently only root-mode agent is supported.
* Integrations cannot be automatically configured yet.
* Windows is not supported.


### Questions

1. How to have a task definition that can be configured to send data to
   different accounts? Can a Service override some parameters of the task
   definition?


## Task definition: Infra agent on ECS Fargate as a sidecar

The script has a Fargate switch (`-f`) that can  be used to enable the Fargate mode. It will create the basic AWS
resources that are required to run Tasks with the sidecar contianer: IAM policies and roles, and the System's Manager
parameter.

In `sidecar_example.json` there is a task definition that contains an nginx container
with the Agent & ECS integration as a sidecar.


### Compatibility

* It should be deployed as a **Task/Service** with the **Fargate Launch Type**.
* The agent is running in forwarder mode, so no host samples are collected.
* Integrations cannot be automatically configured yet.
* Windows is not supported.
