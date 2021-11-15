package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"sync"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
)

type ingestFunc func(ctx context.Context, id string, opts axiom.IngestOptions, events ...axiom.Event) (*axiom.IngestStatus, error)

var logger *zap.Logger

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

type lokiServer struct {
	proxy *httputil.ReverseProxy
	URL   *url.URL
}

func (lk *lokiServer) Host() string {
	return lk.URL.Host
}

func (lk *lokiServer) Scheme() string {
	return lk.URL.Scheme
}

// Multiplexer implements http.Handler.
type Multiplexer struct {
	sync.Mutex
	defaultDataset string
	datasetKey     string
	lokiServer     *lokiServer
	ingestFn       ingestFunc
}

// NewMultiplexer creates a new Multiplexer which uses the passed Axiom client
// to send logs to Axiom.
func NewMultiplexer(ingestFn ingestFunc, lokiEndpoint, defaultDataset, datasetKey string) (*Multiplexer, error) {
	var lk *lokiServer
	if lokiEndpoint != "" {
		hcURL, err := url.Parse(lokiEndpoint)
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(hcURL)
		lk = &lokiServer{proxy: proxy, URL: hcURL}
	}
	return &Multiplexer{
		ingestFn:       ingestFn,
		lokiServer:     lk,
		defaultDataset: defaultDataset,
		datasetKey:     datasetKey,
	}, nil
}

func (m *Multiplexer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if m.lokiServer != nil {
		body := bytes.NewBuffer(nil)
		req.Body = io.NopCloser(io.TeeReader(req.Body, body))
		m.forward(resp, req)
		req.Body = io.NopCloser(body)
	}

	if err := m.multiplex(req); err != nil {
		logger.Error(err.Error())
		if m.lokiServer == nil {
			if _, wErr := resp.Write([]byte(err.Error())); wErr != nil {
				logger.Error(wErr.Error())
			}
		}
	}

	if m.lokiServer == nil {
		if _, wErr := resp.Write([]byte("{}")); wErr != nil {
			logger.Error(wErr.Error())
		}
	}
}

func (m *Multiplexer) forward(resp http.ResponseWriter, req *http.Request) {
	req.URL.Host = m.lokiServer.Host()
	req.URL.Scheme = m.lokiServer.Scheme()
	m.lokiServer.proxy.ServeHTTP(resp, req)
}

func (m *Multiplexer) multiplex(req *http.Request) error {
	defer req.Body.Close()
	if req.Method != "POST" {
		return nil
	}

	var (
		ingestReq *pushRequest
		err       error
	)

	typ := req.Header.Get("Content-Type")
	switch typ {
	case "application/json":
		ingestReq, err = decodeJSONPushRequest(req.Body)
	case "application/x-protobuf":
		ingestReq, err = decodeProtoPushRequest(req.Body)
	default:
		err = fmt.Errorf("unsupported Content-Type %v", typ)
	}

	if err != nil {
		logger.Error(err.Error())
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

		if _, err := m.ingestFn(context.Background(), dataset, axiom.IngestOptions{}, events...); err != nil {
			logger.Error(err.Error())
			return err
		}
		events = make([]axiom.Event, 0)
	}
	return nil
}
