package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/axiomhq/axiom-go/axiom"

	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"
	"github.com/axiomhq/axiom-loki-proxy/version"
)

var (
	deploymentURL = os.Getenv("AXIOM_DEPLOYMENT_URL")
	accessToken   = os.Getenv("AXIOM_ACCESS_TOKEN")
	addr          = flag.String("addr", ":3101", "Listen address <ip>:<port>")
)

func main() {
	log.Print("starting axiom-loki-proxy version", version.Release())

	flag.Parse()

	if deploymentURL == "" {
		log.Fatal("missing AXIOM_DEPLOYMENT_URL")
	}
	if accessToken == "" {
		log.Fatal("missing AXIOM_ACCESS_TOKEN")
	}

	client, err := axiom.NewClient(deploymentURL, accessToken)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/loki/api/v1/push", httpProxy.NewPushHandler(client))

	log.Print("listening on", *addr)

	server := http.Server{Handler: mux, Addr: *addr}
	log.Fatal(server.ListenAndServe())
}
