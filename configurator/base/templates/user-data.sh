#!/bin/bash

# ECS config. Setting this variable will make this node join the ECS cluster defined below.
{
  echo "ECS_CLUSTER=${cluster_name}"
} >> /etc/ecs/ecs.config

start ecs

echo "Done"