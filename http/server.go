package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/axiomhq/axiom-go/axiom"
)

const (
	defaultDataset = "axiom-loki-proxy"
	datasetKey     = "_axiom_dataset"
)

// implements the http.Server interface
type Server struct {
	*http.Server
	axiomClient *axiom.Client
}

func NewServer(port int, client *axiom.Client) *Server {
	srv := &Server{
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", port),
		},
		axiomClient: client,
	}
	http.HandleFunc("/loki/api/v1/push", srv.push)
	return srv
}

func (srv *Server) push(w http.ResponseWriter, r *http.Request) {
	req, err := decodePushRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		client  = srv.axiomClient
		events  = make([]axiom.Event, 0)
		dataset = defaultDataset
	)

	for _, stream := range req.Streams {
		ev := make(axiom.Event)
		labels := stream.Labels.Map()
		if val, ok := labels[datasetKey]; ok {
			dataset = val
			delete(labels, datasetKey)
		}

		for k, v := range labels {
			ev[k] = v
		}

		for _, val := range stream.Entries {
			ev[axiom.TimestampField] = val
			ev["message"] = val
		}

		if res, err := client.Datasets.IngestEvents(context.Background(), dataset, axiom.IngestOptions{}, events...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatalln(err)
		} else {
			log.Println(res)
		}
	}
}
