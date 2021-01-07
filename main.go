package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/axiomhq/axiom-go/axiom"
	httpProxy "github.com/axiomhq/axiom-loki-proxy/http"
)

func dummyIngest(ctx context.Context, id string, opts axiom.IngestOptions, events ...axiom.Event) (*axiom.IngestStatus, error) {
	fmt.Println(events)
	return nil, nil
}

func initHttpPushHandler(mux *http.ServeMux, client *axiom.Client) {
	//handler := httpProxy.NewPushHandler(client)
	handler := &httpProxy.PushHandler{
		IngestFn: dummyIngest,
	}
	mux.Handle("/loki/api/v1/push", handler)
}

func main() {
	var (
		//deploymentURL = os.Getenv("AXM_DEPLOYMENT_URL")
		//accessToken   = os.Getenv("AXM_ACCESS_TOKEN")
		addr = flag.String("addr", "0.0.0.0:3101", "a string <ip>:<port>")
	)

	//client, err := axiom.NewClient(deploymentURL, accessToken)
	//if err != nil {
	//	log.Fatal(err)
	//}

	mux := http.NewServeMux()
	initHttpPushHandler(mux, nil)

	log.Printf("Now listening on %s...\n", *addr)
	server := http.Server{Handler: mux, Addr: *addr}
	log.Fatal(server.ListenAndServe())
}
