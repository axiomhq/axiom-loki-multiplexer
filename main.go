package main

import (
	"context"
	"flag"

	"github.com/axiomhq/axiom-go/axiom"
	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"

	"github.com/axiomhq/pkg/cmd"
	"github.com/axiomhq/pkg/http"
	"go.uber.org/zap"
)

var (
	addr           = flag.String("addr", ":3101", "a string <ip>:<port>")
	lokiURL        = flag.String("loki-addr", ":3100", "a string <ip>:<port>")
	datasetLabel   = flag.String("dataset-label", "_axiom_dataset", "the label key to use as a dataset name")
	defaultDataset = flag.String("default-dataset", "axiom-loki-proxy", "the default dataset to use in case datasetLabel is not found")
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

	mp, err := httpProxy.NewMultiplexer(client, *lokiURL, *datasetLabel, *defaultDataset)
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
