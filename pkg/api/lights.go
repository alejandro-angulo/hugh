package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// TODO: How to enforce valid values?

// CIECoord represents coordinates in the CIE color space.
// The first entry is the x-coordinate and the second entry is the y-coordinate.
// Both values must be between 0 and 1.
type CIECoord [2]float64

// LightState represents the state of a light
type LightState struct {
	On          bool     `json:"on"`
	Brightness  uint8    `json:"bri"`
	Hue         uint16   `json:"hue"`
	Saturation  uint8    `json:"sat"`
	CIECoords   CIECoord `json:"xy"`
	Temperature uint16   `json:"ct"`     // Valid values are 153 (6500K) to 500 (2000K)
	Alert       string   `json:"alert"`  // Valid values are "none", "select", "lselect"
	Effect      string   `json:"effect"` // Valid values are "none" and "colorloop"
	Mode        string   `json:"mode"`
	ColorMode   string   `json:"colormode"`
	Reachable   bool     `json:"reachable"`
	//TransitionTime uint16   `json:"transitiontime"` // Values are multiples of 100ms and default is 4 (400ms)

	// Setting a delta value to 0 stops any ongoing transition

	//BrightnessDelta  int16    `json:"bri_inc"` // Valid values are -254 to 254.
	//SaturationDelta  int16    `json:"sat_inc"` // Valid values are -254 to 254.
	//HueDelta         int16    `json:"hue_inc"` // Valid values are -65534 to 65334.
	//TemperatureDelta int16    `json:"ct_inc"`  // Valid values are -65534 to 65534.
	//CIEDelta         CIECoord `json:"xy_incl"` // TODO: Undertand this better
}

type hueTime struct {
	time.Time
}

func (ht *hueTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	newTime, err := time.Parse("\"2006-01-02T15:04:05\"", string(data))
	if err == nil {
		ht.Time = newTime
	}
	return err
}

// LightSWUpdate holds information about a light's updatability
// TODO: better doc
type LightSWUpdate struct {
	State       string  `json:"state"`
	LastInstall hueTime `json:"lastinstall"`
}

// LightTemperatureRange holds the minimum and maximum temperature values
type LightTemperatureRange struct {
	Minimum uint `json:"min"`
	Maximum uint `json:"max"`
}

// LightColorGamut represents a light's color gamut's RGB coordinates
type LightColorGamut [2]float64

// LightCapabilitiesControl represents information on how a light can be controlled
type LightCapabilitiesControl struct {
	MinimumDim     uint   `json:"mindimlevel"`
	MaxLumen       uint   `json:"maxlumen"`
	ColorGamutType string `json:"colorgamuttype"`

	// Element 0 is the R coordinate. Element 1 is the G coordinate. Element 2 is the B coordinate.
	// TODO: Write helpers to extract R, G, and B?
	ColorGamuts [3]LightColorGamut `json:"colorgamut"`

	TemperatureRange LightTemperatureRange `json:"ct"`
}

// LightStreamingCapabilities holds information about stuff
// TODO: Better doc
type LightStreamingCapabilities struct {
	Renderer bool `json:"renderer"`
	Proxy    bool `json:"proxy"`
}

// LightCapabilities represents information on what a light is capable of
type LightCapabilities struct {
	Certified bool                       `json:"certified"`
	Control   LightCapabilitiesControl   `json:"control"`
	Streaming LightStreamingCapabilities `json:"streaming"`
}

// LightConfig represents a light's configuration
type LightConfig struct {
	Archetype string `json:"archetype"`
	Function  string `json:"function"`
	Direction string `json:"direction"`
}

// Light represents a smart light
type Light struct {
	State        LightState        `json:"state"`
	SWUpdate     LightSWUpdate     `json:"swupdate"`
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	ModelID      string            `json:"modelid"`
	Manufacturer string            `json:"manufacturername"`
	Product      string            `json:"productname"`
	Capabilities LightCapabilities `json:"capabilities"`
	Config       LightConfig       `json:"config"`
	UID          string            `json:"uniqueid"`
	SWVersion    string            `json:"swversion"`
	ID           string            `json:"-"`
	Bridge       *Bridge           `json:"-"`
}

// GetLights retrieves all the lights on a certain bridge
func (bridge *Bridge) GetLights() ([]Light, error) {
	url := fmt.Sprintf("http://%s/api/%s/lights", bridge.IP, bridge.Username)
	resp, err := bridge.API.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]Light
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	lights := []Light{}

	for id, light := range data {
		light.ID = id
		light.Bridge = bridge
		lights = append(lights, light)
	}

	return lights, nil
}
