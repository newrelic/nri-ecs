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
	taskARN := "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	clusterARN := metadata.ClusterARNFromTask(taskARN, "ecs-local-cluster")
	assert.Equal(t, "arn:aws:ecs:us-west-2:111111111111:cluster/ecs-local-cluster", clusterARN)
	otherTaskARN := "arn:aws:ecs:eu-west-1:725889879812:task/af32c116-75e0-4f45-aba1-fc2a203ea5d3"
	otherClusterARN := metadata.ClusterARNFromTask(otherTaskARN, "foobar")
	assert.Equal(t, "arn:aws:ecs:eu-west-1:725889879812:cluster/foobar", otherClusterARN)
}

func TestResourceNameAndARNBase(t *testing.T) {
	taskARN := "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	resourceName, baseARN := metadata.ResourceNameAndARNBase(taskARN)
	assert.Equal(t, "arn:aws:ecs:us-west-2:111111111111", baseARN)
	assert.Equal(t, "ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152", resourceName)
}

func TestMetadataResponse(t *testing.T) {
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
