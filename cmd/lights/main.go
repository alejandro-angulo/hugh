package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/alejandro-angulo/hugh/pkg/api"
)

func main() {
	var flagTimeoutSeconds int
	var flagUsername string
	var flagAddress string

	flag.IntVar(&flagTimeoutSeconds, "timeout", 3, "Timeout in seconds for web requests")
	flag.StringVar(&flagUsername, "username", "", "Username to use for web requests")
	flag.StringVar(&flagAddress, "address", "", "Address of the Bridge to connect to")

	flag.Parse()

	if flagUsername == "" {
		log.Fatalln("Username must be supplied using the username flag.")
	}

	IP := net.ParseIP(flagAddress)
	if IP == nil {
		log.Fatalln("A valid IP address must be supplied with the address flag.")
	}

	client := http.Client{
		Timeout: time.Duration(flagTimeoutSeconds) * time.Second,
	}

	bridge := api.Bridge{
		IP:       IP,
		Username: flagUsername,
	}

	apiObj := api.API{
		Client:         client,
		TimeoutSeconds: flagTimeoutSeconds,
	}

	lights, err := apiObj.GetLights(bridge)
	if err != nil {
		log.Fatal(err)
	}

	for _, light := range lights {
		fmt.Printf("%s\n", light.Name)
	}
}
