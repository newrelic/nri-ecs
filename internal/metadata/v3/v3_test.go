package v3

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterARNFromTask(t *testing.T) {
	taskARN := "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	clusterARN := ClusterARNFromTask(taskARN, "ecs-local-cluster")
	assert.Equal(t, "arn:aws:ecs:us-west-2:111111111111:cluster/ecs-local-cluster", clusterARN)
	otherTaskARN := "arn:aws:ecs:eu-west-1:725889879812:task/af32c116-75e0-4f45-aba1-fc2a203ea5d3"
	otherClusterARN := ClusterARNFromTask(otherTaskARN, "foobar")
	assert.Equal(t, "arn:aws:ecs:eu-west-1:725889879812:cluster/foobar", otherClusterARN)
}

func TestResourceNameAndARNBase(t *testing.T) {
	taskARN := "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	resourceName, baseARN := ResourceNameAndARNBase(taskARN)
	assert.Equal(t, "arn:aws:ecs:us-west-2:111111111111", baseARN)
	assert.Equal(t, "ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152", resourceName)

}

func TestMetadataResponse(t *testing.T) {
	var taskResponseJSONFile = "testdata/task_response.json"
	taskJSON, err := ioutil.ReadFile(taskResponseJSONFile)
	assert.NoError(t, err)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(taskJSON)
	}))
	defer testServer.Close()

	client := &http.Client{}
	response, err := MetadataResponse(client, testServer.URL)
	assert.NoError(t, err)

	assert.Equal(t, taskJSON, response)
}
