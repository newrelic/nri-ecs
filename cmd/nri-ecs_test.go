package main_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newrelic/infra-integrations-sdk/integration"
	main "github.com/newrelic/nri-ecs/cmd"
	"github.com/newrelic/nri-ecs/cmd/testdata"
	"github.com/newrelic/nri-ecs/internal/ecs/metadata"
)

const v3 = `{
	"Cluster": "ecs-local-cluster", 
	"TaskARN": "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	}`

const v4 = `{
	"Cluster": "ecs-local-cluster", 
	"TaskARN": "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152",
	"LaunchType": "EC2"
	}`

func TestIntegrationPublish(t *testing.T) {
	args := main.ArgumentList{}
	// Capture Stdout to get the integration output
	stdout := os.Stdout
	readerOut, writerOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writerOut
	defer func() {
		os.Stdout = stdout
	}()

	cases := map[string]struct {
		response       string
		endpointEnvVar string
	}{
		"from_v3_metadata_endpoint": {
			response:       v3,
			endpointEnvVar: metadata.ContainerMetadataEnvVar,
		},
		"from_v4_metadata_endpoint": {
			response:       v4,
			endpointEnvVar: metadata.ContainerMetadataV4EnvVar,
		},
	}
	for testCaseName, testData := range cases {
		testData := testData

		t.Run(testCaseName, func(t *testing.T) {
			ts := metadataServer(t, testData.response)
			t.Setenv(testData.endpointEnvVar, ts.URL)

			ecsIntegration, _ := integration.New("com.newrelic.ecs", "0.0.0")

			assert.NoError(t, main.Run(ecsIntegration, args))

			// Read the integration output from the captured stdout
			b := make([]byte, 1024)
			n, err := readerOut.Read(b)
			if err != nil {
				t.Fatal(err)
			}
			var result interface{}
			err = json.Unmarshal(b[:n], &result)
			require.NoError(t, err)

			var expected interface{}
			err = json.Unmarshal([]byte(testdata.IntegrationOutput), &expected)
			require.NoError(t, err)

			assert.Equal(t, expected, result)
		})
	}
}

func metadataServer(t *testing.T, response string) *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/task", r.RequestURI)

			_, err := w.Write([]byte(response))
			require.NoError(t, err)
		},
	))

	t.Cleanup(func() {
		testServer.Close()
	})

	return testServer
}
