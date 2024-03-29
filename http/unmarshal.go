package http

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/loghttp"
	pb "github.com/grafana/loki/pkg/logproto"
)

type Stream struct {
	Entries []pb.Entry
	Labels  loghttp.LabelSet
}

type PushRequest struct {
	Streams []Stream
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

func DecodeProtoPushRequest(r io.Reader) (*PushRequest, error) {
	var req pb.PushRequest

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

	ret := &PushRequest{
		Streams: make([]Stream, len(req.Streams)),
	}

	for i, s := range req.Streams {
		labels, err := convertLabelsString(s.Labels)
		if err != nil {
			return nil, err
		}

		ret.Streams[i] = Stream{
			Labels:  loghttp.LabelSet(labels),
			Entries: s.Entries,
		}
	}

	return ret, nil
}

func DecodeJSONPushRequest(b io.Reader) (*PushRequest, error) {
	var req loghttp.PushRequest
	if err := json.NewDecoder(b).Decode(&req); err != nil {
		return nil, err
	}

	ret := &PushRequest{
		Streams: make([]Stream, len(req.Streams)),
	}

	for i, s := range req.Streams {
		ret.Streams[i] = newStream(s)
	}

	return ret, nil
}

func newStream(s *loghttp.Stream) Stream {
	ret := Stream{
		Entries: make([]pb.Entry, len(s.Entries)),
		Labels:  s.Labels,
	}

	for i, e := range s.Entries {
		ret.Entries[i] = pb.Entry{
			Timestamp: e.Timestamp,
			Line:      e.Line,
		}
	}

	return ret
}
