package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alejandro-angulo/hugh/pkg/api"
	"github.com/grandcat/zeroconf"
)

func main() {
	var flagTimeoutSeconds int
	flag.IntVar(&flagTimeoutSeconds, "timeout", 10, "Timeout in seconds for web requests")
	flag.Parse()

	client := http.Client{
		Timeout: time.Duration(flagTimeoutSeconds) * time.Second,
	}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}
	apiObj := api.API{
		Client:         client,
		Browser:        resolver,
		TimeoutSeconds: flagTimeoutSeconds,
	}

	bridges, err := apiObj.Discover()
	if err != nil {
		log.Fatal(err)
	}

	var bridge api.Bridge
	numBridges := len(bridges)

	if numBridges == 0 {
		log.Fatal("No Hue Bridges found!")
	} else if numBridges == 1 {
		bridge = bridges[0]
	} else {
		fmt.Println("More than one Hue Bridge discovered. Choose one of the following:")
		for i, candidate := range bridges {
			fmt.Printf("[%d] %v (ID: `%s` Model: `%s`)\n", i, candidate.IP, candidate.ID, candidate.Model)
		}

		fmt.Print("Enter selection number: ")
		var selection int
		fmt.Scanln(&selection)

		if selection < 0 || selection >= numBridges {
			log.Fatalf("Invalid selection `%d` (must be greatenr than 0 and less then %d)", selection, numBridges)
		}

		bridge = bridges[selection]
	}

	fmt.Println("Attempting to associate with bridge. Please press button on your bridge.")
	fmt.Println("Press the Enter Key when ready...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	username, err := bridge.Connect()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(username)
}
