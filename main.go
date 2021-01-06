package main

import (
	"flag"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-loki-proxy/http"
)

func main() {
	var (
		deploymentURL = os.Getenv("AXM_DEPLOYMENT_URL")
		accessToken   = os.Getenv("AXM_ACCESS_TOKEN")
		port          = flag.Int("port", 3101, "an int")
	)

	client, err := axiom.NewClient(deploymentURL, accessToken)
	if err != nil {
		log.Fatal(err)
	}

	srv := http.NewServer(*port, client)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
