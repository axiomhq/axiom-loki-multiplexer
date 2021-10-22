package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
)

var logger *zap.Logger

func Decompress(rdr io.ReadCloser, typ string) (*pushRequest, error) {
	switch typ {
	case "application/json":
		return decodeJsonPushRequest(rdr)
	case "application/x-protobuf":
		return decodeProtoPushRequest(rdr)
	default:
		return nil, fmt.Errorf("unsupported Content-Type %v", typ)
	}
}

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// type check on axiom.Event incase it's ever not a map[string]interface{}
	// so we can use unsafe.Pointer for a quick type conversion instead of allocating a new slice
	if !reflect.TypeOf(axiom.Event{}).ConvertibleTo(reflect.TypeOf(map[string]interface{}{})) {
		panic("axiom.Event is not a map[string]interface{}, please contact support")
	}
}

type Multiplexer struct {
	client         *axiom.Client
	proxy          *httputil.ReverseProxy
	lokiURL        *url.URL
	datasetLabel   string
	defaultDataset string
}

func NewMultiplexer(client *axiom.Client, lokiEndpoint, datasetLabel, defaultDataset string) (*Multiplexer, error) {
	lokiURL, err := url.Parse(lokiEndpoint)
	if err != nil {
		return nil, err
	}
	return &Multiplexer{
		client:         client,
		proxy:          httputil.NewSingleHostReverseProxy(lokiURL),
		lokiURL:        lokiURL,
		datasetLabel:   datasetLabel,
		defaultDataset: defaultDataset,
	}, nil
}

func (m *Multiplexer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	body := bytes.NewBuffer(nil)
	req.Body = io.NopCloser(io.TeeReader(req.Body, body))
	m.forward(resp, req)
	req.Body = io.NopCloser(body)
	if err := m.multiplex(req); err != nil {
		logger.Error(err.Error())
	}
}

func (m *Multiplexer) forward(resp http.ResponseWriter, req *http.Request) {
	req.URL.Host = m.lokiURL.Host
	req.URL.Scheme = m.lokiURL.Scheme
	m.proxy.ServeHTTP(resp, req)
}

func (m *Multiplexer) multiplex(req *http.Request) error {
	if req.Method != "POST" {
		return nil
	}

	switch {
	case strings.HasPrefix(req.URL.Path, "/loki/api/v1/push"):
		pReq, err := Decompress(req.Body, req.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		defer req.Body.Close()

		events := m.streamsToEvents(pReq)
		if err != nil {
			return err
		}
		for dataset, events := range events {
			if err := m.sendEvents(req.Context(), dataset, events...); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported path %v", req.URL.Path)
	}

	return nil
}

func (m *Multiplexer) streamsToEvents(pReq *pushRequest) map[string][]axiom.Event {
	var (
		collections = make(map[string][]axiom.Event, len(pReq.Streams))
		dataset     = m.defaultDataset
	)

	for _, stream := range pReq.Streams {
		labels := stream.Labels.Map()
		if val, ok := labels[m.datasetLabel]; ok {
			dataset = val
			delete(labels, m.datasetLabel)
		}

		events := make([]axiom.Event, len(stream.Entries))
		for _, val := range stream.Entries {
			ev := make(axiom.Event)
			for k, v := range labels {
				ev[k] = v
			}
			ev[axiom.TimestampField] = val.Timestamp
			ev["message"] = val.Line
			events = append(events, ev)
		}
		collections[dataset] = events
	}
	return collections
}

func (m *Multiplexer) sendEvents(ctx context.Context, dataset string, events ...axiom.Event) error {
	opts := axiom.IngestOptions{}

	status, err := m.client.Datasets.IngestEvents(ctx, dataset, opts, events...)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(status); err != nil {
		return err
	}

	logger.Info(buf.String())
	return nil
}
