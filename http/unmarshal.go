package http

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/golang/snappy"
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

func convertLabelsString(str string) (map[string]string, error) {
	str = strings.Replace(str, ", ", ", \"", -1)
	str = strings.Replace(str, "{", "{\"", -1)
	str = strings.Replace(str, "=", "\":", -1)
	labels := map[string]string{}
	err := json.Unmarshal([]byte(str), &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func decodeProtoPushRequest(r io.Reader) (*pushRequest, error) {
	var req logproto.PushRequest

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = snappy.Decode(nil, b)
	if err != nil {
		return nil, err
	}
	if err := req.Unmarshal(b); err != nil {
		return nil, err
	}

	ret := &pushRequest{
		Streams: make([]stream, len(req.Streams)),
	}

	for i, s := range req.Streams {
		labels, err := convertLabelsString(s.Labels)
		if err != nil {
			return nil, err
		}

		ret.Streams[i] = stream{
			Labels:  loghttp.LabelSet(labels),
			Entries: s.Entries,
		}
	}

	return ret, nil
}

func decodeJsonPushRequest(b io.Reader) (*pushRequest, error) {
	var req loghttp.PushRequest
	if err := json.NewDecoder(b).Decode(&req); err != nil {
		return nil, err
	}

	ret := &pushRequest{
		Streams: make([]stream, len(req.Streams)),
	}

	for i, s := range req.Streams {
		ret.Streams[i] = newStream(s)
	}

	return ret, nil
}

func newStream(s *loghttp.Stream) stream {
	ret := stream{
		Entries: make([]logproto.Entry, len(s.Entries)),
		Labels:  s.Labels,
	}

	for i, e := range s.Entries {
		ret.Entries[i] = logproto.Entry{
			Timestamp: e.Timestamp,
			Line:      e.Line,
		}
	}

	return ret
}
