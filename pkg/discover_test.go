package hugh

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grandcat/zeroconf"
)

type BrowseFunc func(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error

func (f BrowseFunc) Browse(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
	return f(ctx, service, domain, entries)
}

func NewTestAPI(fn BrowseFunc) *API {
	return &API{
		Browser:        BrowseFunc(fn),
		TimeoutSeconds: 1,
	}
}

type testData struct {
	text string
	IP   []byte
}

func TestDiscover(t *testing.T) {
	tests := []struct {
		name       string
		bridges    []Bridge
		bridgeData []testData
	}{
		{
			name: "Test bridges are found",
			bridges: []Bridge{
				Bridge{
					ID:    "test",
					Model: "foo",
					IP:    []byte{127, 0, 0, 1},
				},
				Bridge{
					ID:    "foobar",
					Model: "bar",
					IP:    []byte{192, 168, 1, 66},
				},
			},
			bridgeData: []testData{
				testData{
					text: "[bridgeid=test modelid=foo]",
					IP:   []byte{127, 0, 0, 1},
				},
				testData{
					text: "[bridgeid=foobar modelid=bar]",
					IP:   []byte{192, 168, 1, 66},
				},
			},
		},
		{
			name:    "Test no bridge is found",
			bridges: []Bridge{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := NewTestAPI(func(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
				for _, data := range tt.bridgeData {
					entries <- &zeroconf.ServiceEntry{
						Text:     []string{data.text},
						AddrIPv4: []net.IP{data.IP},
					}
				}
				return nil
			})

			got, _ := api.Discover()

			assertBridges(t, got, tt.bridges)
		})
	}

	t.Run("Test error is returned on mDNS browse failure", func(t *testing.T) {
		expectedError := errors.New("Simulated failure")

		api := NewTestAPI(func(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
			return expectedError
		})

		got, err := api.Discover()

		if got != nil {
			t.Errorf("Expected nil for bridges but got %v", got)
		}

		if err != expectedError {
			t.Errorf("Expected %v but got %v", expectedError, err)
		}
	})
}

func assertBridges(t *testing.T, got, want []Bridge) {

	numGot := len(got)
	numWant := len(want)

	if numGot != numWant {
		t.Errorf("got %d bridges, want %d bridges", numGot, numWant)
	}

	for i, actualBridge := range got {
		expectedBridge := want[i]
		if !cmp.Equal(actualBridge, expectedBridge) {
			t.Errorf("Bridge #%d mismatch, got %q want %q", i, actualBridge, expectedBridge)
		}
	}
}
