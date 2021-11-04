package metadata_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
	"github.com/stretchr/testify/assert"
)

func TestClusterARNFromTask(t *testing.T) {
	t.Parallel()

	taskARN := "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	clusterARN := metadata.ClusterARNFromTask(taskARN, "ecs-local-cluster")
	assert.Equal(t, "arn:aws:ecs:us-west-2:111111111111:cluster/ecs-local-cluster", clusterARN)
	otherTaskARN := "arn:aws:ecs:eu-west-1:725889879812:task/af32c116-75e0-4f45-aba1-fc2a203ea5d3"
	otherClusterARN := metadata.ClusterARNFromTask(otherTaskARN, "foobar")
	assert.Equal(t, "arn:aws:ecs:eu-west-1:725889879812:cluster/foobar", otherClusterARN)
}

func TestResourceNameAndARNBase(t *testing.T) {
	t.Parallel()

	taskARN := "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	resourceName, baseARN := metadata.ResourceNameAndARNBase(taskARN)
	assert.Equal(t, "arn:aws:ecs:us-west-2:111111111111", baseARN)
	assert.Equal(t, "ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152", resourceName)
}

func TestMetadataResponse(t *testing.T) {
	t.Parallel()

	taskResponseJSONFile := "testdata/task_response.json"
	taskJSON, err := ioutil.ReadFile(taskResponseJSONFile)
	assert.NoError(t, err)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(taskJSON)
	}))
	defer testServer.Close()

	client := &http.Client{}
	response, err := metadata.MetadataResponse(client, testServer.URL)
	assert.NoError(t, err)

	assert.Equal(t, taskJSON, response)
}

func TestClusterToClusterName(t *testing.T) {
	t.Parallel()

	tt := []struct {
		cluster, expectedClusterName string
	}{
		{"my-cluster", "my-cluster"},
		{"arn:aws:ecs:eu-south-1337:1:cluster/my_awesome_long_cluster_name_which_nobody_ever_used_before", "my_awesome_long_cluster_name_which_nobody_ever_used_before"},
		{"arn:aws:ecs:eu-south-1337:1:cluster/x", "x"},
		{"arn:aws:ecs:eu-south-1337:1:cluster/", "arn:aws:ecs:eu-south-1337:1:cluster/"},
		{"", ""},
		{"hello world", "hello world"},
		{"arn:aws:iam::123456789012:user/Development/product_1234/askdk", "arn:aws:iam::123456789012:user/Development/product_1234/askdk"}, // not an ECS arn
		{"arn:aws:ecs:eu-south-1337:cluster/fsi", "arn:aws:ecs:eu-south-1337:cluster/fsi"},                                                 // missing account number
	}

	for _, testCase := range tt {
		clusterName := metadata.ClusterToClusterName(testCase.cluster)
		assert.Equal(t, testCase.expectedClusterName, clusterName)
	}
}

func TestTaskMetadataEndpoint(t *testing.T) {
	v4 := "http://localhost/v4"
	v3 := "http://localhost/v3"
	t.Run("returns_v4_endpoint_if_v3_and_v4_exists", func(t *testing.T) {
		t.Setenv(metadata.ContainerMetadataV4EnvVar, v4)
		t.Setenv(metadata.ContainerMetadataEnvVar, v3)

		endpoint, found := metadata.TaskMetadataEndpoint()

		assert.Equal(t, v4+"/task", endpoint)
		assert.Equal(t, true, found)
	})

	t.Run("returns_v3_if_only_v3_exists", func(t *testing.T) {
		t.Setenv(metadata.ContainerMetadataEnvVar, v3)

		endpoint, found := metadata.TaskMetadataEndpoint()

		assert.Equal(t, v3+"/task", endpoint)
		assert.Equal(t, true, found)
	})

	t.Run("returns_v4_if_only_v4_exists", func(t *testing.T) {
		t.Setenv(metadata.ContainerMetadataEnvVar, v4)

		endpoint, found := metadata.TaskMetadataEndpoint()

		assert.Equal(t, v4+"/task", endpoint)
		assert.Equal(t, true, found)
	})

	t.Run("found_no_endpoints_if_no_envars_exists", func(t *testing.T) {
		endpoint, found := metadata.TaskMetadataEndpoint()

		assert.Equal(t, "", endpoint)
		assert.Equal(t, false, found)
	})
}

func TestFargateLaunchType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		isFargate  bool
		launchType string
		expected   string
	}{
		{
			isFargate:  true,
			launchType: "",
			expected:   metadata.EcsFargateLaunchType,
		},
		{
			isFargate:  false,
			launchType: "",
			expected:   metadata.EcsEC2LaunchType,
		},
		{
			isFargate:  false,
			launchType: "EC2",
			expected:   metadata.EcsEC2LaunchType,
		},
		{
			isFargate:  false,
			launchType: "FARGATE",
			expected:   metadata.EcsFargateLaunchType,
		},
		{
			isFargate:  false,
			launchType: "EXTERNAL",
			expected:   metadata.EcsExternalLaunchType,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.launchType, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, metadata.LaunchType(testCase.isFargate, testCase.launchType))
		})
	}
}
