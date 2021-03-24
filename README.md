# Axiom Loki Proxy

[![Go Workflow][go_workflow_badge]][go_workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]
[![Docker][docker_badge]][docker]

---

## Table of Contents

1. [Introduction](#introduction)
1. [Usage](#usage)
1. [Contributing](#contributing)
1. [License](#license)

## Introduction

_Axiom Loki Proxy_ ships logs to Axiom, using [Loki HTTP API][1].

  [1]: https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push

## Installation

### Download the pre-compiled and archived binary manually

Binary releases are available on [GitHub Releases][2].

  [2]: https://github.com/axiomhq/axiom-loki-proxy/releases/latest

### Install using [Homebrew](https://brew.sh)

```shell
$ brew tap axiomhq/tap
$ brew install axiom-loki-proxy
```

To update:

```shell
$ brew update
$ brew upgrade axiom-loki-proxy
```

### Install using `go get`

```shell
$ go get -u github.com/axiomhq/axiom-loki-proxy/cmd/axiom-loki-proxy
```

### Install from source

```shell
$ git clone https://github.com/axiomhq/axiom-loki-proxy.git
$ cd axiom-loki-proxy
$ make build
```

### Run the Docker image

Docker images are available on [DockerHub][docker].

## Usage

1. Set the following environment variables:
   * `AXIOM_DEPLOYMENT_URL`: URL of the Axiom deployment to use
   * `AXIOM_ACCESS_TOKEN`: **Personal Access** or **Ingest** token. Can be
     created under `Profile` or `Settings > Ingest Tokens`. For security reasons
     it is advised to use an Ingest Token with minimal privileges only.

2. Run it: `./axiom-loki-proxy` or using docker:
   `docker run -p3101:3101/tcp axiomhq/axiom-loki-proxy`

## Contributing

Feel free to submit PRs or to fill issues. Every kind of help is appreciated. 

Before committing, `make` should run without any issues.

Kindly check our [Contributing](Contributing.md) guide on how to propose
bugfixes and improvements, and submitting pull requests to the project.

## License

&copy; Axiom, Inc., 2021

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

<!-- Badges -->

[go_workflow]: https://github.com/axiomhq/axiom-loki-proxy/actions?query=workflow%3Ago
[go_workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/axiom-loki-proxy/go?style=flat-square&ghcache=unused
[coverage]: https://codecov.io/gh/axiomhq/axiom-loki-proxy
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/axiom-loki-proxy.svg?style=flat-square&ghcache=unused
[report]: https://goreportcard.com/report/github.com/axiomhq/axiom-loki-proxy
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/axiom-loki-proxy?style=flat-square&ghcache=unused
[release]: https://github.com/axiomhq/axiom-loki-proxy/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-loki-proxy.svg?style=flat-square&ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-loki-proxy.svg?color=blue&style=flat-square&ghcache=unused
[docker]: https://hub.docker.com/r/axiomhq/axiom-loki-proxy
[docker_badge]: https://img.shields.io/docker/pulls/axiomhq/axiom-loki-proxy.svg?style=flat-square&ghcache=unused
