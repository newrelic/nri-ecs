# ecs-cluster-configurator

This Terraform repository uses [Workspaces](https://www.terraform.io/docs/state/workspaces.html) to create multiple clusters. 

## Add a new clusters

Run the following commands to create a new cluster, with 5 worker nodes of type t3.large:

```shell
# create a new workspace
terraform workspace new my-cluster-name

# deploy
terraform apply -var 'instance_type=t3.large worker_nodes=3'
```

## Modifying a cluster

This example shows how to scale the amount of worker nodes to 5, using the cluster we created before. Do not forget to repeat all configuration variables again.

```shell
# select the workspace
terraform workspace select my-cluster-name

# apply changes 
terraform apply -var 'instance_type=t3.large worker_nodes=5'
```

## Delete a cluster

Run the following commands to delete a cluster:

```shell
# select the workspace
terraform workspace select my-cluster-name

# execute order 66
terraform destroy
```

## Seeing available clusters

Run the following command to see list of available clusters:

```shell
terraform workspace list
```
