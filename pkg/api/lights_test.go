package api

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGetLights(t *testing.T) {
	tests := []struct {
		name      string
		lights    []Light
		lightData string
	}{
		{
			name:      "Test no lights are found",
			lights:    []Light{},
			lightData: "{}",
		},
		{
			name: "Test light is found",
			lights: []Light{
				Light{
					State: LightState{
						On:          false,
						Brightness:  1,
						Hue:         33761,
						Saturation:  254,
						Effect:      "none",
						CIECoords:   [2]float64{0.3171, 0.3366},
						Temperature: 159,
						Alert:       "none",
						ColorMode:   "xy",
						Mode:        "homeautomation",
						Reachable:   true,
					},
					SWUpdate: LightSWUpdate{
						State:       "noupdates",
						LastInstall: hueTime{time.Date(2018, 1, 2, 19, 24, 20, 0, time.UTC)},
					},
					Type:         "Extended color light",
					Name:         "Hue color lamp 7",
					ModelID:      "LCT007",
					Manufacturer: "Philips",
					Product:      "Hue color lamp",
					Capabilities: LightCapabilities{
						Certified: true,
						Control: LightCapabilitiesControl{
							MinimumDim:     5000,
							MaxLumen:       600,
							ColorGamutType: "B",
							ColorGamuts: [3]LightColorGamut{
								[2]float64{0.675, 0.322},
								[2]float64{0.409, 0.518},
								[2]float64{0.167, 0.040},
							},
							TemperatureRange: LightTemperatureRange{
								Maximum: 500,
								Minimum: 153,
							},
						},
						Streaming: LightStreamingCapabilities{
							Renderer: true,
							Proxy:    false,
						},
					},
					Config: LightConfig{
						Archetype: "sultanbulb",
						Function:  "mixed",
						Direction: "omnidirectional",
					},
					UID:       "00:17:88:01:00:bd:c7:b9-0b",
					SWVersion: "5.105.0.21169",
					ID:        "1",
					// Bridge: Set in test since this is dynamically created
				},
			},
			lightData: `{
"1": {
        "state": {
            "on": false,
            "bri": 1,
            "hue": 33761,
            "sat": 254,
            "effect": "none",
            "xy": [
                0.3171,
                0.3366
            ],
            "ct": 159,
            "alert": "none",
            "colormode": "xy",
            "mode": "homeautomation",
            "reachable": true
        },
        "swupdate": {
            "state": "noupdates",
            "lastinstall": "2018-01-02T19:24:20"
        },
        "type": "Extended color light",
        "name": "Hue color lamp 7",
        "modelid": "LCT007",
        "manufacturername": "Philips",
        "productname": "Hue color lamp",
        "capabilities": {
            "certified": true,
            "control": {
                "mindimlevel": 5000,
                "maxlumen": 600,
                "colorgamuttype": "B",
                "colorgamut": [
                    [
                        0.675,
                        0.322
                    ],
                    [
                        0.409,
                        0.518
                    ],
                    [
                        0.167,
                        0.04
                    ]
                ],
                "ct": {
                    "min": 153,
                    "max": 500
                }
            },
            "streaming": {
                "renderer": true,
                "proxy": false
            }
        },
        "config": {
            "archetype": "sultanbulb",
            "function": "mixed",
            "direction": "omnidirectional"
        },
        "uniqueid": "00:17:88:01:00:bd:c7:b9-0b",
        "swversion": "5.105.0.21169"
    }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := NewTestAPI(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(tt.lightData))),
				}, nil
			}, DefaultBrowse)

			bridge := Bridge{
				IP:  []byte{127, 0, 0, 1},
				API: api,
			}

			for i := range tt.lights {
				tt.lights[i].Bridge = &bridge
			}

			lights, err := bridge.GetLights()
			if err != nil {
				t.Errorf("Expected no error but got %v", err)
			}

			assertLights(t, lights, tt.lights)
		})
	}
}

func assertLights(t *testing.T, got, want []Light) {
	t.Helper()

	numGot := len(got)
	numWant := len(want)

	if numGot != numWant {
		t.Errorf("got %d lights, want %d lights", numGot, numWant)
	}

	for i, actualLight := range got {
		expectedLight := want[i]

		diff := cmp.Diff(actualLight, expectedLight)
		if diff != "" {
			log.Println(diff)
			t.Errorf("Light #%d mismatch", i)
		}
	}
}
