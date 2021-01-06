package http

import (
	"encoding/json"
	"io"

	"github.com/grafana/loki/pkg/loghttp"
	"github.com/grafana/loki/pkg/logproto"
)

type stream struct {
	Entries []logproto.Entry
	Labels  loghttp.LabelSet
}

type pushRequest struct {
	Streams []stream
}

func decodePushRequest(b io.Reader) (*pushRequest, error) {
	var request loghttp.PushRequest
	if err := json.NewDecoder(b).Decode(&request); err != nil {
		return nil, err
	}
	return newPushRequest(request), nil
}

func newPushRequest(r loghttp.PushRequest) *pushRequest {
	ret := &pushRequest{
		Streams: make([]stream, len(r.Streams)),
	}

	for i, s := range r.Streams {
		ret.Streams[i] = newStream(s)
	}

	return ret
}

func newStream(s *loghttp.Stream) stream {
	ret := stream{
		Entries: make([]logproto.Entry, len(s.Entries)),
		Labels:  s.Labels,
	}

	for i, e := range s.Entries {
		ret.Entries[i] = newEntry(e)
	}

	return ret
}

func newEntry(e loghttp.Entry) logproto.Entry {
	return logproto.Entry{
		Timestamp: e.Timestamp,
		Line:      e.Line,
	}
}
