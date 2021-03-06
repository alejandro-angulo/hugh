package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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
	// HTTP client
	Client http.Client

	// Multicast DNS browser
	Browser MulticastBrowser

	// The number of seconds to wait for a Hue Bridge to be discovered
	TimeoutSeconds int

	// TODO: Keep a slice of (pointers to) Bridges?
}

// APIError represents the error dictionary from the Connect response
type APIError struct {
	Type        int    `json:"type"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// ConnectSuccess represents the success dictionary from the Connect response
type ConnectSuccess struct {
	Username string `json:"username"`
}

// ConnectResponse represents the API's response structure
type ConnectResponse struct {
	Error   APIError       `json:"error"`
	Success ConnectSuccess `json:"success"`
}

// Bridge represents a Phillips Hue bridge
type Bridge struct {
	API   *API   `json:"-"`
	ID    string `json:"id"`
	Model string `json:"model"`
	IP    net.IP `json:"internalipaddress"` // TODO: Store v4 and v6 IPs as separate attributes?

	// The username to use when communicating with this brige
	Username string `json:"-"`
}

// ErrBodyLengthTooLong is returned when the API's body is longer than expected
// The API returns arrays for some responses even when only one item is expected
var ErrBodyLengthTooLong = errors.New("Response body length longer than expected")

func parseServiceEntryText(entry *zeroconf.ServiceEntry) map[string]string {
	var data = map[string]string{}

	for _, record := range entry.Text {
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
			newBridge := Bridge{
				API:   api,
				ID:    textData["bridgeid"],
				Model: textData["modelid"],
				IP:    entry.AddrIPv4[0], // Assume first item in slice is what we want

			}
			bridges = append(bridges, newBridge)
			fmt.Printf("Found Hue bridge at %s (ID: `%s` Model: `%s`)\n", newBridge.IP.String(), newBridge.ID, newBridge.Model)
		}
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
// Returns the user ID  and sets the Bridge's Username attribute if sucessful
func (bridge *Bridge) Connect() (string, error) {
	url := fmt.Sprintf("http://%s/api", bridge.IP.String())

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(map[string]string{
		"devicetype": fmt.Sprintf("hugh#%s", hostname),
	})
	if err != nil {
		return "", err
	}

	resp, err := bridge.API.Client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body := []ConnectResponse{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", err
	}

	// Only one item is expected (even though the API returns an array)
	if len(body) != 1 {
		return "", ErrBodyLengthTooLong
	}

	respData := body[0]
	username := respData.Success.Username

	if username == "" {
		return "", fmt.Errorf(
			"Failed to associate with bridge. Error{type:`%d`, address:`%s`, description:`%s`}",
			respData.Error.Type,
			respData.Error.Address,
			respData.Error.Description,
		)
	}

	bridge.Username = respData.Success.Username

	return respData.Success.Username, nil
}
