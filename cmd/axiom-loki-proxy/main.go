package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/pkg/version"

	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"
)

var (
	deploymentURL = os.Getenv("AXIOM_URL")
	accessToken   = os.Getenv("AXIOM_TOKEN")
	addr          = flag.String("addr", ":8080", "Listen address <ip>:<port>")
)

func main() {
	log.Print("starting axiom-loki-proxy version ", version.Release())

	flag.Parse()

	if deploymentURL == "" {
		deploymentURL = axiom.CloudURL
	}
	if accessToken == "" {
		log.Fatal("missing AXIOM_TOKEN")
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
