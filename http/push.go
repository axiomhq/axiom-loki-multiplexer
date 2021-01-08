package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/axiomhq/axiom-go/axiom"
)

const (
	defaultDataset = "axiom-loki-proxy"
	datasetKey     = "_axiom_dataset"
)

type ingestFunc func(ctx context.Context, id string, opts axiom.IngestOptions, events ...axiom.Event) (*axiom.IngestStatus, error)

// PushHandler implements http.Handler.
type PushHandler struct {
	sync.Mutex
	ingestFn ingestFunc
}

// NewPushHandler creates a new PushHandler which uses the passed Axiom client
// to send logs to Axiom.
func NewPushHandler(client *axiom.Client) *PushHandler {
	return &PushHandler{
		ingestFn: client.Datasets.IngestEvents,
	}
}

// ServeHTTP implements http.Handler.
func (push *PushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	push.Lock()
	defer push.Unlock()

	var (
		req *pushRequest
		err error
	)

	typ := r.Header.Get("Content-Type")
	switch typ {
	case "application/json":
		req, err = decodeJSONPushRequest(r.Body)
	case "application/x-protobuf":
		req, err = decodeProtoPushRequest(r.Body)
	default:
		err = fmt.Errorf("unsupported Content-Type %v", typ)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		events  = make([]axiom.Event, 0)
		dataset = defaultDataset
	)

	for _, stream := range req.Streams {
		labels := stream.Labels.Map()
		if val, ok := labels[datasetKey]; ok {
			dataset = val
			delete(labels, datasetKey)
		}

		for _, val := range stream.Entries {
			ev := make(axiom.Event)
			for k, v := range labels {
				ev[k] = v
			}
			ev[axiom.TimestampField] = val.Timestamp
			ev["message"] = val.Line
			events = append(events, ev)
		}

		if _, err := push.ingestFn(context.Background(), dataset, axiom.IngestOptions{}, events...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatalln(err)
		}
		events = make([]axiom.Event, 0)
	}
}
