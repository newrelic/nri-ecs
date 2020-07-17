package ecs_test

import (
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.datanerd.us/fsi/nri-ecs/pkg/ecs"
)

const taskResponseJSONFile = "testdata/task_response.json"

/*

	Be aware, the tests in this package will not work when using Go version 1.11 or above.
	The integrations SDK parses flags, and the testing framework does so as well...

	To run these tests, please use go 1.11:

	brew install go@1.11
	/usr/local/opt/go@1.11/bin/go test ./...
*/

// skipIfGoVersionIsAbove11 does some things we can better not talk about
func skipIfGoVersionIsAbove11(t *testing.T) {

	// version looks like go1.13.7 or go1.13
	versions := strings.Split(runtime.Version(), ".")
	major, err := strconv.Atoi(versions[1])
	if err != nil {
		return
	}

	if major > 11   {
		t.Skip("Warning! Skipping test because go version is not compatible (requires go version <= 11)")
	}
}

func TestNewClusterEntity(t *testing.T) {
	skipIfGoVersionIsAbove11(t)

	integration, _ := ecs.NewIntegration(&ecs.Args)
	cluster, err := ecs.NewClusterEntity("arn:aws:ecs:us-west-2:xxxxxxxx:cluster/ecs-local-cluster", integration)
	assert.NoError(t, err)
	assert.Equal(t, "cluster/ecs-local-cluster", cluster.Metadata.Name)
	assert.Equal(t, "arn:aws:ecs:us-west-2:xxxxxxxx", cluster.Metadata.Namespace)
}

func TestAddClusterInventory(t *testing.T) {
	skipIfGoVersionIsAbove11(t)

	integration, _ := ecs.NewIntegration(&ecs.Args)

	entity, err := integration.Entity("foo", "bar")
	assert.NoError(t, err)

	err = ecs.AddClusterInventory("clusterName", "clusterARN", entity)
	assert.NoError(t, err)

	item, ok := entity.Inventory.Item("cluster")
	assert.True(t, ok, "inventory not found")
	assert.Equal(t, "clusterName", item["name"])
	assert.Equal(t, "clusterARN", item["name"])
}

func TestAddClusterInventoryToLocalEntity(t *testing.T) {
	skipIfGoVersionIsAbove11(t)

	i, _ := ecs.NewIntegration(&ecs.Args)

	ecsClusterName := "my-cluster"
	ecsClusterARN := "arn:my-cluster"
	awsRegion := "us-east-1"
	launchType := ecs.NewLaunchType(true)

	err := ecs.AddClusterInventoryToLocalEntity(ecsClusterName, ecsClusterARN, awsRegion, launchType, i)
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
	skipIfGoVersionIsAbove11(t)

	integration, _ := ecs.NewIntegration(&ecs.Args)

	entity, err := integration.Entity("foo", "bar")
	assert.NoError(t, err)

	metricSet, err := ecs.NewClusterHeartbeatMetricSet(
		"ecs-local-cluster",
		"arn:cluster:ecs-local-cluster",
		entity,
	)
	assert.NoError(t, err)

	assert.Equal(t, "ecs-local-cluster", metricSet.Metrics["clusterName"])
	assert.Equal(t, "arn:cluster:ecs-local-cluster", metricSet.Metrics["arn"])
}
