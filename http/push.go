package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
)

type ingestFunc func(ctx context.Context, id string, opts axiom.IngestOptions, events ...axiom.Event) (*axiom.IngestStatus, error)

type lokiServer struct {
	multiplexer *httputil.ReverseProxy
	URL         *url.URL
}

func (lk *lokiServer) Host() string {
	return lk.URL.Host
}

func (lk *lokiServer) Scheme() string {
	return lk.URL.Scheme
}

// Multiplexer implements `http.Handler`.
type Multiplexer struct {
	sync.Mutex

	defaultDataset string
	datasetKey     string

	logger     *zap.Logger
	lokiServer *lokiServer
	ingestFn   ingestFunc
}

// NewMultiplexer creates a new Multiplexer which uses the passed Axiom client
// to send logs to Axiom.
func NewMultiplexer(logger *zap.Logger, ingestFn ingestFunc, lokiEndpoint, defaultDataset, datasetKey string) (*Multiplexer, error) {
	var lk *lokiServer
	if lokiEndpoint != "" {
		hcURL, err := url.Parse(lokiEndpoint)
		if err != nil {
			return nil, err
		}
		multiplexer := httputil.NewSingleHostReverseProxy(hcURL)
		lk = &lokiServer{multiplexer: multiplexer, URL: hcURL}
	}
	return &Multiplexer{
		logger:         logger,
		ingestFn:       ingestFn,
		lokiServer:     lk,
		defaultDataset: defaultDataset,
		datasetKey:     datasetKey,
	}, nil
}

// ServeHTTP implements `http.Handler`.
func (m *Multiplexer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if m.lokiServer != nil {
		body := bytes.NewBuffer(nil)
		req.Body = io.NopCloser(io.TeeReader(req.Body, body))
		m.forward(resp, req)
		req.Body = io.NopCloser(body)
	}

	if err := m.multiplex(req); err != nil {
		m.logger.Error(err.Error())
		if m.lokiServer == nil {
			if _, wErr := resp.Write([]byte(err.Error())); wErr != nil {
				m.logger.Error(wErr.Error())
			}
		}
	}

	if m.lokiServer == nil {
		if _, wErr := resp.Write([]byte("{}")); wErr != nil {
			m.logger.Error(wErr.Error())
		}
	}
}

func (m *Multiplexer) forward(resp http.ResponseWriter, req *http.Request) {
	req.URL.Host = m.lokiServer.Host()
	req.URL.Scheme = m.lokiServer.Scheme()
	m.lokiServer.multiplexer.ServeHTTP(resp, req)
}

func (m *Multiplexer) multiplex(req *http.Request) error {
	if req.Method != http.MethodPost {
		return nil
	}

	var (
		ingestReq *PushRequest
		err       error
	)

	typ := req.Header.Get("Content-Type")
	switch typ {
	case "application/json":
		ingestReq, err = DecodeJSONPushRequest(req.Body)
	case "application/x-protobuf":
		ingestReq, err = DecodeProtoPushRequest(req.Body)
	default:
		err = fmt.Errorf("unsupported Content-Type %v", typ)
	}

	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	var (
		events  = make([]axiom.Event, 0)
		dataset = m.defaultDataset
	)

	for _, stream := range ingestReq.Streams {
		labels := stream.Labels.Map()
		if val, ok := labels[m.datasetKey]; ok {
			dataset = val
			delete(labels, m.datasetKey)
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

		if _, err := m.ingestFn(req.Context(), dataset, axiom.IngestOptions{}, events...); err != nil {
			m.logger.Error(err.Error())
			return err
		}
		events = make([]axiom.Event, 0)
	}
	return nil
}
