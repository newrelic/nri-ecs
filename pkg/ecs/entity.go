package ecs

import (
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	v3 "source.datanerd.us/fsi/nri-ecs/internal/metadata/v3"
)

const (
	ecsClusterEntityType = "aws:ecs:cluster"
	ecsClusterEventType  = "EcsClusterSample"
)

func NewClusterEntity(clusterARN string, i *integration.Integration) (*integration.Entity, error) {
	clusterName, arnPrefix := v3.ResourceNameAndARNBase(clusterARN)
	clusterEntity, err := i.Entity("cluster/"+clusterName, arnPrefix)
	if err != nil {
		return nil, err
	}
	return clusterEntity, nil
}

// AddClusterInventoryLocalEntity adds some ecs attributes as inventory
// to the integration's local entity.
func AddClusterInventoryToLocalEntity(clusterName, clusterARN, awsRegion string, launchType LaunchType, integration *integration.Integration) error {

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

func AddClusterInventory(clusterName, clusterARN string, entity *integration.Entity) error {
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

func NewClusterHeartbeatMetricSet(clusterName string, clusterARN string, entity *integration.Entity) (*metric.Set, error) {
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
