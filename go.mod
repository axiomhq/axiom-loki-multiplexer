module github.com/axiomhq/axiom-loki-proxy

go 1.16

require (
	github.com/axiomhq/axiom-go v0.0.0-20210312122006-3294c6b958f9
	github.com/axiomhq/pkg v0.0.0-20210318171555-dc26762456be
	github.com/golang/snappy v0.0.3
	github.com/golangci/golangci-lint v1.38.0
	github.com/goreleaser/goreleaser v0.159.0
	github.com/grafana/loki v1.6.1
	github.com/stretchr/testify v1.7.0
	gotest.tools/gotestsum v1.6.2
)

replace k8s.io/client-go => k8s.io/client-go v0.20.1
