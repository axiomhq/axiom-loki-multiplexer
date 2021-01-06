package http

import (
	"context"
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

// implements the http.Server interface
type PushHandler struct {
	sync.Mutex
	ingestFn ingestFunc
}

func NewPushHandler(client *axiom.Client) *PushHandler {
	return &PushHandler{
		ingestFn: client.Datasets.IngestEvents,
	}
}

func (push *PushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	push.Lock()
	defer push.Unlock()

	req, err := decodePushRequest(r.Body)

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

		if res, err := push.ingestFn(context.Background(), dataset, axiom.IngestOptions{}, events...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatalln(err)
		} else {
			log.Println(res)
		}
		events = make([]axiom.Event, 0)

	}
}
