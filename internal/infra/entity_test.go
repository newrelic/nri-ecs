package infra_test

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
	"github.com/newrelic/nri-ecs/internal/infra"
)

func TestNewClusterEntity(t *testing.T) {
	i, _ := integration.New("test", "dev")
	cluster, err := infra.NewClusterEntity("arn:aws:ecs:us-west-2:xxxxxxxx:cluster/ecs-local-cluster", i)
	assert.NoError(t, err)
	assert.Equal(t, "cluster/ecs-local-cluster", cluster.Metadata.Name)
	assert.Equal(t, "arn:aws:ecs:us-west-2:xxxxxxxx", cluster.Metadata.Namespace)
}

func TestAddClusterInventory(t *testing.T) {
	i, _ := integration.New("test", "dev")

	entity, err := i.Entity("foo", "bar")
	assert.NoError(t, err)

	err = infra.AddClusterInventory("clusterName", "clusterARN", entity)
	assert.NoError(t, err)

	item, ok := entity.Inventory.Item("cluster")
	assert.True(t, ok, "inventory not found")
	assert.Equal(t, "clusterName", item["name"])
	assert.Equal(t, "clusterARN", item["arn"])
}

func TestAddClusterInventoryToLocalEntity(t *testing.T) {
	i, _ := integration.New("test", "dev")

	ecsClusterName := "my-cluster"
	ecsClusterARN := "arn:my-cluster"
	awsRegion := "us-east-1"
	launchType := metadata.LaunchType(true, "")

	err := infra.AddClusterInventoryToLocalEntity(ecsClusterName, ecsClusterARN, awsRegion, launchType, i)
	require.NoError(t, err)

	e := i.LocalEntity()
	ecsCluster, ok := e.Inventory.Item("host")

	assert.True(t, ok, "inventory not found")

	expected := inventory.Item(map[string]interface{}{
		"ecsClusterName": ecsClusterName,
		"ecsClusterArn":  ecsClusterARN,
		"awsRegion":      awsRegion,
		"ecsLaunchType":  launchType,
	})
	assert.Equal(t, expected, ecsCluster)
}

func TestNewClusterHeartbeatMetricSet(t *testing.T) {
	integration, _ := integration.New("test", "dev")

	entity, err := integration.Entity("foo", "bar")
	assert.NoError(t, err)

	metricSet, err := infra.NewClusterHeartbeatMetricSet(
		"ecs-local-cluster",
		"arn:cluster:ecs-local-cluster",
		entity,
	)
	assert.NoError(t, err)

	assert.Equal(t, "ecs-local-cluster", metricSet.Metrics["clusterName"])
	assert.Equal(t, "arn:cluster:ecs-local-cluster", metricSet.Metrics["arn"])
}
