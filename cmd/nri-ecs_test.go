package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newrelic/nri-ecs/cmd/testdata"
)

func TestIntegrationPublish(t *testing.T) {
	clusterName := "ecs-local-cluster"
	// Mock the metadata endpoint
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/task", r.RequestURI)
			response := fmt.Sprintf(
				`{"Cluster": "%s", "TaskARN": "arn:aws:ecs:us-west-2:111111111111:task/%s/37e873f6-37b4-42a7-af47-eac7275c6152"}`,
				clusterName,
				clusterName,
			)
			w.Write([]byte(response))
		},
	))
	defer ts.Close()
	os.Setenv("ECS_CONTAINER_METADATA_URI", ts.URL)
	defer os.Clearenv()

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

	main()

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
}
