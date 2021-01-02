package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grandcat/zeroconf"
)

var ErrShouldNotBeCalled = errors.New("Function called unexpectedly")

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (x RoundTripFunc) Equal(y RoundTripFunc) bool {
	return reflect.ValueOf(x).Pointer() == reflect.ValueOf(y).Pointer()
}

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func DefaultRoundTrip(*http.Request) (*http.Response, error) {
	return nil, ErrShouldNotBeCalled
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

type BrowseFunc func(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error

func (x BrowseFunc) Equal(y BrowseFunc) bool {
	return reflect.ValueOf(x).Pointer() == reflect.ValueOf(y).Pointer()
}

func (f BrowseFunc) Browse(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
	return f(ctx, service, domain, entries)
}

func DefaultBrowse(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
	return ErrShouldNotBeCalled
}

func NewTestAPI(transportFn RoundTripFunc, browseFn BrowseFunc) *API {
	return &API{
		Browser:        BrowseFunc(browseFn),
		Client:         *NewTestClient(transportFn),
		TimeoutSeconds: 1,
	}
}

type testData struct {
	text []string
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
					text: []string{"bridgeid=test", "modelid=foo"},
					IP:   []byte{127, 0, 0, 1},
				},
				testData{
					text: []string{"bridgeid=foobar", "modelid=bar"},
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
			api := NewTestAPI(
				RoundTripFunc(DefaultRoundTrip),
				func(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
					for _, data := range tt.bridgeData {
						entries <- &zeroconf.ServiceEntry{
							Text:     data.text,
							AddrIPv4: []net.IP{data.IP},
						}
					}
					return nil
				},
			)

			for i := range tt.bridges {
				tt.bridges[i].API = api
			}

			got, err := api.Discover()

			if err != nil {
				t.Errorf("Expected no error but got %v", err)
			}

			assertBridges(t, got, tt.bridges)
		})
	}

	t.Run("Test error is returned on mDNS browse failure", func(t *testing.T) {
		expectedError := errors.New("Simulated failure")

		api := NewTestAPI(
			DefaultRoundTrip,
			func(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
				return expectedError
			},
		)

		got, err := api.Discover()

		if got != nil {
			t.Errorf("Expected nil for bridges but got %v", got)
		}

		if err != expectedError {
			t.Errorf("Expected %v but got %v", expectedError, err)
		}
	})
}

func testConnect(t *testing.T) {
	t.Run("Test associating with Hue Bridge fails", func(t *testing.T) {
		api := NewTestAPI(func(*http.Request) (*http.Response, error) {
			json := `[{"error": {"type": 101, "address: "", "description": ""}}]`

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
			}, nil
		}, DefaultBrowse)

		bridge := Bridge{
			IP:  []byte{127, 0, 0, 1},
			API: api,
		}

		username, err := bridge.Connect()

		if username != "" {
			t.Errorf("Expected no username to be returned but got %s", username)
		}

		if err == nil {
			t.Error("Expected an error to have been returned")
		}
	})

	t.Run("Test associating with Hue Bridge is successful", func(t *testing.T) {
		expectedUsername := "testUser"

		api := NewTestAPI(func(*http.Request) (*http.Response, error) {
			json := fmt.Sprintf(`[{"success": {"username": "%s"}}]`, expectedUsername)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
			}, nil
		}, DefaultBrowse)

		bridge := Bridge{
			IP:  []byte{127, 0, 0, 1},
			API: api,
		}

		username, err := bridge.Connect()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if username != expectedUsername {
			t.Errorf("Expected username to be %s but got %s", expectedUsername, username)
		}

		if username != bridge.Username {
			t.Errorf("Bridge username was not correctly updated. Saw %s but got %s", bridge.Username, username)
		}
	})
}

func assertBridges(t *testing.T, got, want []Bridge) {
	t.Helper()

	numGot := len(got)
	numWant := len(want)

	if numGot != numWant {
		t.Errorf("got %d bridges, want %d bridges", numGot, numWant)
	}

	for i, actualBridge := range got {
		expectedBridge := want[i]

		diff := cmp.Diff(actualBridge, expectedBridge)
		if diff != "" {
			log.Println(diff)
			t.Errorf("Bridge #%d mismatch", i)
		}
	}
}
