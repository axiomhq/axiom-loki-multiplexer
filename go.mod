module github.com/axiomhq/axiom-loki-proxy

go 1.15

require (
	github.com/axiomhq/axiom-go v0.0.0-20201215212509-678033418d51
	github.com/golang/snappy v0.0.2
	github.com/golangci/golangci-lint v1.35.2
	github.com/goreleaser/goreleaser v0.154.0
	github.com/grafana/loki v1.6.1
	github.com/stretchr/testify v1.6.1
	gotest.tools/gotestsum v0.6.0
)

replace k8s.io/client-go => k8s.io/client-go v0.20.1
