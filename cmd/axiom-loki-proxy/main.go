package main

import (
	"context"
	"flag"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/pkg/cmd"
	"github.com/axiomhq/pkg/http"
	"go.uber.org/zap"

	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"
)

var (
	addr           = flag.String("addr", ":8080", "Listen address <ip>:<port>")
	lokiURL        = flag.String("loki-url", "http://localhost:3100", "Loki URL")
	byPassLoki     = flag.Bool("bypass", false, "Bypass Loki")
	defaultDataset = flag.String("default-dataset", "axiom-loki-proxy", "Default dataset")
	datasetKey     = flag.String("dataset-key", "_axiom_dataset_key", "Dataset key")
)

func main() {
	cmd.Run("axiom-loki-proxy", run,
		cmd.WithValidateAxiomCredentials(),
	)
}

func run(ctx context.Context, log *zap.Logger, client *axiom.Client) error {
	// Export `AXIOM_TOKEN` and `AXIOM_ORG_ID` for Axiom Cloud.
	// Export `AXIOM_URL` and `AXIOM_TOKEN` for Axiom Selfhost.

	flag.Parse()

	url := ""
	if !*byPassLoki {
		url = *lokiURL
	}

	mp, err := httpProxy.NewMultiplexer(client.Datasets.IngestEvents, url, *defaultDataset, *datasetKey)
	if err != nil {
		return cmd.Error("create multiplexer", err)
	}

	srv, err := http.NewServer(*addr, mp,
		http.WithBaseContext(ctx),
		http.WithLogger(log),
	)
	if err != nil {
		return cmd.Error("create http server", err)
	}
	defer func() {
		if shutdownErr := srv.Shutdown(); shutdownErr != nil {
			log.Error("stopping server", zap.Error(shutdownErr))
			return
		}
	}()

	srv.Run(ctx)

	log.Info("server listening",
		zap.String("address", srv.ListenAddr().String()),
		zap.String("network", srv.ListenAddr().Network()),
	)

	select {
	case <-ctx.Done():
		log.Warn("received interrupt, exiting gracefully")
	case err := <-srv.ListenError():
		return cmd.Error("error starting http server, exiting gracefully", err)
	}

	return nil
}
