package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alejandro-angulo/hugh/pkg/api"
	"github.com/grandcat/zeroconf"
)

func main() {
	var flagTimeoutSeconds int
	flag.IntVar(&flagTimeoutSeconds, "timeout", 3, "Timeout in seconds for web requests")
	flag.Parse()

	client := http.Client{
		Timeout: time.Duration(3) * time.Second,
	}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}
	api := api.API{
		Client:         client,
		Browser:        resolver,
		TimeoutSeconds: flagTimeoutSeconds,
	}

	bridges, err := api.Discover()
	if err != nil {
		log.Fatal(err)
	}

	bridge := bridges[0]
	username, err := api.Connect(bridge)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(username)
}
