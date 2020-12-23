package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/grandcat/zeroconf"
)

type Bridge struct {
	ID string `json:"id"`
	IP string `json:"internalipaddress"`
}

func discover() (bridges []Bridge) {
	resp, err := http.Get("https://discovery.meethue.com/")
	if err != nil {
		log.Fatal(err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Received status code %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(string(content))

	err = json.Unmarshal(content, &bridges)
	if err != nil {
		log.Fatal(err)
	}

	return bridges
}

func connect(bridge Bridge) {
	url := fmt.Sprintf("http://%s/api", bridge.IP)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	payload, err := json.Marshal(map[string]string{
		"devicetype": fmt.Sprintf("hugh#%s", hostname),
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var results []map[string]map[string]string
	json.NewDecoder(resp.Body).Decode(&results)

	if resp.StatusCode != 200 {
		log.Fatalf("Received status code %d", resp.StatusCode)
	}

	value := results[0]

	if val, ok := value["error"]; ok {
		log.Println(fmt.Sprint(val))
	}
}

func main() {
	service := "_hue._tcp"
	domain := "local"
	waitTime := 10

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	log.Println("Scanning network for Hue bridges...")

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println(entry.Text)
		}
		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(waitTime))
	defer cancel()
	err = resolver.Browse(ctx, service, domain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()

	// TODO: Should discover return errors or log.Fatal (currently done)
	//bridges := discover()

	//num_bridges := len(bridges)

	//fmt.Printf("Number of bridges found: %d\n", num_bridges)

	//var bridge Bridge

	//if num_bridges == 0 {
	//log.Fatal("No Hue Bridges found!")
	//} else if num_bridges == 1 {
	//// TODO: Use this bridge
	//bridge = bridges[0]
	//} else {
	//// TODO: Let user select bridge
	//}

	//fmt.Printf("%s %s\n", bridge.ID, bridge.IP)
	//connect(bridge)
}
