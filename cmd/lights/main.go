package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/alejandro-angulo/hugh/pkg/api"
	"github.com/rivo/tview"
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

	apiObj := api.API{
		Client:         client,
		TimeoutSeconds: flagTimeoutSeconds,
	}

	bridge := api.Bridge{
		IP:       IP,
		Username: flagUsername,
		API:      &apiObj,
	}

	lights, err := bridge.GetLights()
	if err != nil {
		log.Fatal(err)
	}

	app := tview.NewApplication()
	finder(app, lights)
	if err := app.Run(); err != nil {
		fmt.Println(string(debug.Stack()))
		log.Fatal(err)
	}
}

func finder(app *tview.Application, lights []api.Light) {
	lightsList := tview.NewList()
	lightsList.SetBorder(true).SetTitle("Lights")

	lightInfo := tview.NewList()
	lightInfo.SetBorder(true).SetTitle("Light Info")
	lightInfo.SetDoneFunc(func() {
		app.SetFocus(lightsList)
		lightInfo.Clear()
	})

	flex := tview.NewFlex().
		AddItem(lightsList, 0, 1, true).
		AddItem(lightInfo, 0, 1, false)

	getLightStatus := func(light *api.Light) (status string) {
		status = "Off"
		if light.State.On {
			status = "On"
		}

		return
	}

	addLightToList := func(light api.Light) {
		lightsList.AddItem(light.Name, "", 0, func() {
			lightInfo.Clear()

			lightInfo.AddItem("Active", getLightStatus(&light), 'a', func() {
				light.ToggleLight()
				lightInfo.SetItemText(0, "Active", getLightStatus(&light))
			})

			lightInfo.AddItem("Brightness", fmt.Sprintf("%v", light.State.Brightness), 'b', nil)
			lightInfo.AddItem("Hue", fmt.Sprintf("%v", light.State.Hue), 'h', nil)
			lightInfo.AddItem("Saturation", fmt.Sprintf("%v", light.State.Saturation), 's', nil)

			app.SetFocus(lightInfo)
		})
	}

	for _, light := range lights {
		addLightToList(light)
	}

	pages := tview.NewPages().
		AddPage("someName", flex, true, true)
	app.SetRoot(pages, true)
}
