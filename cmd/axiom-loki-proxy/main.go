package main

import (
	"context"
	"flag"
	stdHttp "net/http"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/pkg/cmd"
	"github.com/axiomhq/pkg/http"
	"go.uber.org/zap"

	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"
)

var addr = flag.String("addr", ":8080", "Listen address <ip>:<port>")

func main() {
	cmd.Run("axiom-loki-proxy", run,
		cmd.WithValidateAxiomCredentials(),
	)
}

func run(ctx context.Context, log *zap.Logger, client *axiom.Client) error {
	// Export `AXIOM_TOKEN` and `AXIOM_ORG_ID` for Axiom Cloud.
	// Export `AXIOM_URL` and `AXIOM_TOKEN` for Axiom Selfhost.

	flag.Parse()

	mux := stdHttp.NewServeMux()
	mux.Handle("/loki/api/v1/push", httpProxy.NewPushHandler(client))

	srv, err := http.NewServer(*addr, mux,
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
