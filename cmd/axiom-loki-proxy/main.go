package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	xhttp "github.com/axiomhq/pkg/http"
	"github.com/axiomhq/pkg/version"

	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"
)

const (
	exitOK int = iota
	exitConfig
	exitInternal
)

var addr = flag.String("addr", ":8080", "Listen address <ip>:<port>")

func main() {
	os.Exit(Main())
}

func Main() int {
	// Export `AXIOM_TOKEN` and `AXIOM_ORG_ID` for Axiom Cloud
	// Export `AXIOM_URL` and `AXIOM_TOKEN` for Axiom Selfhost

	log.Print("starting axiom-loki-proxy version ", version.Release())

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer cancel()

	client, err := axiom.NewClient()
	if err != nil {
		log.Print(err)
		return exitConfig
	} else if err = client.ValidateCredentials(ctx); err != nil {
		log.Print(err)
		return exitConfig
	}

	mux := http.NewServeMux()
	mux.Handle("/loki/api/v1/push", httpProxy.NewPushHandler(client))

	srv, err := xhttp.NewServer(*addr, mux)
	if err != nil {
		log.Print(err)
		return exitInternal
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*5)
		defer shutdownCancel()

		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			log.Print(shutdownErr)
		}
	}()

	srv.Run(ctx)

	log.Print("listening on ", srv.ListenAddr().String())

	select {
	case <-ctx.Done():
		log.Print("received interrupt, exiting gracefully")
	case err := <-srv.ListenError():
		log.Print("error starting http server, exiting gracefully: ", err)
		return exitInternal
	}

	return exitOK
}
