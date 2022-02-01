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

func Test_PopulateIntegration(t *testing.T) {
	i, _ := integration.New("test", "dev")
	taskMetadata := metadata.TaskResponse{
		Cluster:    "ecs-local-cluster",
		TaskARN:    "arn:aws:ecs:us-west-2:xxxxxxxx:cluster/ecs-local-cluster",
		LaunchType: metadata.EcsEC2LaunchType,
	}

	t.Run("when_clusterMetadata_is_complete", func(t *testing.T) {
		t.Parallel()

		clusterMetadata := infra.NewClusterMetadata(taskMetadata, true)
		assert.NoError(t, infra.PopulateIntegration(i, clusterMetadata))

		// The extra entity is the LocalEntity.
		require.Len(t, i.Entities, 2)
		clusterEntity := i.Entities[0]

		t.Run("generates_the_cluster_entity", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "cluster/ecs-local-cluster", i.Entities[0].Metadata.Name)
			assert.Equal(t, "arn:aws:ecs:us-west-2:xxxxxxxx", i.Entities[0].Metadata.Namespace)
		})

		t.Run("generates_inventory", func(t *testing.T) {
			t.Parallel()

			item, ok := clusterEntity.Inventory.Item("cluster")

			assert.True(t, ok, "inventory not found")
			assert.Equal(t, clusterMetadata.Name, item["name"])
			assert.Equal(t, clusterMetadata.ARN, item["arn"])
		})

		t.Run("add_cluster_metadata_to_local_entity", func(t *testing.T) {
			t.Parallel()

			e := i.LocalEntity()

			ecsCluster, ok := e.Inventory.Item("host")
			assert.True(t, ok, "inventory not found")

			expected := inventory.Item(map[string]interface{}{
				"ecsClusterName": clusterMetadata.Name,
				"ecsClusterArn":  clusterMetadata.ARN,
				"awsRegion":      clusterMetadata.Region,
				"ecsLaunchType":  clusterMetadata.LaunchType,
			})

			assert.Equal(t, expected, ecsCluster)
		})

		t.Run("add_heartbeat_metric_set", func(t *testing.T) {
			t.Parallel()

			assert.Len(t, clusterEntity.Metrics, 1)

			metrics := clusterEntity.Metrics[0].Metrics
			require.Len(t, metrics, 3)

			assert.Equal(t, clusterMetadata.Name, metrics["clusterName"])
			assert.Equal(t, clusterMetadata.ARN, metrics["arn"])
			assert.Equal(t, infra.EcsClusterEventType, metrics["event_type"])
		})
	})
}

func Test_PopulateIntegrationFargate(t *testing.T) {
	i, _ := integration.New("test", "dev")
	taskMetadata := metadata.TaskResponse{
		Cluster:    "ecs-local-cluster",
		TaskARN:    "arn:aws:ecs:us-west-2:xxxxxxxx:cluster/ecs-local-cluster",
		LaunchType: metadata.EcsFargateLaunchType,
	}

	t.Run("when_clusterMetadata_is_complete_and_is_fargate", func(t *testing.T) {
		t.Parallel()

		clusterMetadata := infra.NewClusterMetadata(taskMetadata, true)
		assert.NoError(t, infra.PopulateIntegration(i, clusterMetadata))

		// The extra entity is the LocalEntity.
		require.Len(t, i.Entities, 2)
		clusterEntity := i.Entities[0]

		t.Run("add_heartbeat_metric_set_has_all_metrics", func(t *testing.T) {
			t.Parallel()

			assert.Len(t, clusterEntity.Metrics, 1)

			metrics := clusterEntity.Metrics[0].Metrics
			require.Len(t, metrics, 7)

			assert.Equal(t, clusterMetadata.Name, metrics["clusterName"])
			assert.Equal(t, clusterMetadata.Name, metrics["ecsClusterName"])
			assert.Equal(t, clusterMetadata.LaunchType, metrics["ecsLaunchType"])
			assert.Equal(t, clusterMetadata.Region, metrics["awsRegion"])
			assert.Equal(t, clusterMetadata.ARN, metrics["ecsClusterArn"])
			assert.Equal(t, clusterMetadata.ARN, metrics["arn"])
			assert.Equal(t, infra.EcsClusterEventType, metrics["event_type"])
		})
	})
}
