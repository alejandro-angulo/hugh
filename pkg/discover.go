package hugh

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// MulticastBrowser defines the interface required for the mDNS client
type MulticastBrowser interface {
	// Browse finds mDNS services of a give type in a given domain.
	Browse(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error
}

// API represents the Phillips Hue API
type API struct {
	Browser MulticastBrowser

	// The number of seconds to wait for a Hue Bridge to be discovered
	TimeoutSeconds int
}

// Bridge represents a Phillips Hue bridge
type Bridge struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	IP    net.IP `json:"internalipaddress"` // TODO: Store v4 and v6 IPs as separate attributes?
}

func parseServiceEntryText(entry *zeroconf.ServiceEntry) map[string]string {
	var data = map[string]string{}

	text := entry.Text[0]
	records := strings.Fields(text[1 : len(text)-1])

	for _, record := range records {
		rawData := strings.Split(record, "=")
		data[rawData[0]] = rawData[1]
	}

	return data
}

// Discover searches for Phillips Hue bridges on the local network using mDNS
func (api *API) Discover() ([]Bridge, error) {
	var bridges []Bridge
	log.Println("Scanning network for Hue bridges...")

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			textData := parseServiceEntryText(entry)
			bridges = append(bridges, Bridge{
				ID:    textData["bridgeid"],
				Model: textData["modelid"],
				IP:    entry.AddrIPv4[0], // Assume first item in slice is what we want
			})
		}
		log.Println("No more entries.")
	}(entries)

	waitTime := time.Second * time.Duration(api.TimeoutSeconds)
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()
	err := api.Browser.Browse(ctx, "_hue._tcp", "local", entries)
	if err != nil {
		return nil, err
	}

	<-ctx.Done()

	return bridges, nil
}

// Connect associates with a Phillips Hue Bridge
func (api *API) Connect() {}
