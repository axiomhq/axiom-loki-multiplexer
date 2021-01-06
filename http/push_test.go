package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/stretchr/testify/assert"
)

func sampleIngest(ctx context.Context, id string, opts axiom.IngestOptions, events ...axiom.Event) (*axiom.IngestStatus, error) {
	fmt.Println(events)
	return nil, nil
}

var sampleStreams = map[string]interface{}{
	"streams": []map[string]interface{}{
		{
			"stream": map[string]string{
				"label1": "value1",
				"label2": "value2",
			},
			"values": [][2]string{
				{"1", "hello world"},
				{"2", "the answer is 42"},
				{"3", "foobar"},
			},
		},
	},
}

func TestMyHandler(t *testing.T) {
	push := &PushHandler{
		ingestFn: sampleIngest,
	}

	server := httptest.NewServer(push)
	defer server.Close()

	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(sampleStreams)
	assert.NoError(t, err)

	resp, err := http.Post(server.URL, "application/json", buf)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.EqualValues(t, resp.StatusCode, 200)
	assert.NoError(t, resp.Body.Close())
}
