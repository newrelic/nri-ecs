package infra

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"

	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
)

const (
	ecsClusterEventType = "EcsClusterSample"
)

type ClusterMetadata struct {
	Name       string
	ARN        string
	Region     string
	LaunchType string
}

func NewClusterMetadata(taskMetadata metadata.TaskResponse, fargate bool) ClusterMetadata {
	clusterName := metadata.ClusterToClusterName(taskMetadata.Cluster)

	return ClusterMetadata{
		Name:       clusterName,
		Region:     metadata.AWSRegionFromTask(taskMetadata.TaskARN),
		LaunchType: metadata.LaunchType(fargate, taskMetadata.LaunchType),
		ARN:        metadata.ClusterARNFromTask(taskMetadata.TaskARN, clusterName),
	}
}

func PopulateIntegration(i *integration.Integration, cm ClusterMetadata) error {
	clusterEntity, err := newClusterEntity(cm.ARN, i)
	if err != nil {
		return fmt.Errorf("unable to create cluster entity: %v", err)
	}

	// Generate the Cluster Entity through the inventory hack.
	if err = addClusterInventory(cm.Name, cm.ARN, clusterEntity); err != nil {
		log.Error("unable to register cluster inventory: %v", err)
	}

	if _, err = newClusterHeartbeatMetricSet(cm.Name, cm.ARN, clusterEntity); err != nil {
		log.Error("unable to create metrics for cluster: %v", err)
	}

	// This allows all samples and metrics that are sent by the agent to be aggregated with this metadata.
	if err = addClusterMetadataToLocalEntityInventory(cm.Name, cm.ARN, cm.Region, cm.LaunchType, i); err != nil {
		log.Error("unable to register cluster inventory to local entity: %v", err)
	}

	return nil
}

func newClusterEntity(clusterARN string, i *integration.Integration) (*integration.Entity, error) {
	clusterName, arnPrefix := metadata.ResourceNameAndARNBase(clusterARN)
	clusterEntity, err := i.Entity("cluster/"+clusterName, arnPrefix)
	if err != nil {
		return nil, err
	}
	return clusterEntity, nil
}

func addClusterMetadataToLocalEntityInventory(clusterName, clusterARN, awsRegion string, launchType string, integration *integration.Integration) error {
	entity := integration.LocalEntity()
	err := entity.SetInventoryItem("host", "ecsClusterName", clusterName)
	if err != nil {
		return err
	}

	err = entity.SetInventoryItem("host", "ecsClusterArn", clusterARN)
	if err != nil {
		return err
	}

	err = entity.SetInventoryItem("host", "awsRegion", awsRegion)
	if err != nil {
		return err
	}

	err = entity.SetInventoryItem("host", "ecsLaunchType", launchType)
	if err != nil {
		return err
	}

	return nil
}

func addClusterInventory(clusterName, clusterARN string, entity *integration.Entity) error {
	err := entity.SetInventoryItem("cluster", "name", clusterName)
	if err != nil {
		return err
	}

	err = entity.SetInventoryItem("cluster", "arn", clusterARN)
	if err != nil {
		return err
	}

	return nil
}

func newClusterHeartbeatMetricSet(clusterName string, clusterARN string, entity *integration.Entity) (*metric.Set, error) {
	metricSet := entity.NewMetricSet(ecsClusterEventType)
	var err error

	err = metricSet.SetMetric("clusterName", clusterName, metric.ATTRIBUTE)
	if err != nil {
		return nil, err
	}

	err = metricSet.SetMetric("arn", clusterARN, metric.ATTRIBUTE)
	if err != nil {
		return nil, err
	}

	return metricSet, nil
}
