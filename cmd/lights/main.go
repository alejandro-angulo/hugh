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

	addLightToList := func(i int, light api.Light) {
		lightsList.AddItem(light.Name, "", 0, func() {
			lightInfo.Clear()

			lightInfo.AddItem("Active", getLightStatus(&light), 'a', func() {
				light.ToggleLight()
				lightInfo.SetItemText(i, "Active", getLightStatus(&light))
			})

			lightInfo.AddItem("Brightness", fmt.Sprintf("%v", light.State.Brightness), 'b', nil)
			lightInfo.AddItem("Hue", fmt.Sprintf("%v", light.State.Hue), 'h', nil)
			lightInfo.AddItem("Saturation", fmt.Sprintf("%v", light.State.Saturation), 's', nil)

			app.SetFocus(lightInfo)
		})
	}

	for i, light := range lights {
		addLightToList(i, light)
	}

	pages := tview.NewPages().
		AddPage("someName", flex, true, true)
	app.SetRoot(pages, true)

	//root := tview.NewTreeNode("Lights")
	//tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)

	//for _, light := range lights {
	//// fmt.Printf("[%d] %s\n", i, light.Name)
	//lightNode := tview.NewTreeNode(light.Name).SetReference(light).SetSelectable(true).SetExpanded(false)

	//lightStatus := getLightStatus(&light)
	//lightStatusNode := tview.NewTreeNode(lightStatus).SetReference(light)
	//lightStatusNode.SetSelectedFunc(func() {
	////light := node.GetReference()
	//// TODO: handle exceptional case when light is none
	//light.ToggleLight()
	//lightStatusNode.SetText(getLightStatus(&light))
	//})
	//lightNode.AddChild(lightStatusNode)

	//root.AddChild(lightNode)
	//}

	//tree.SetSelectedFunc(func(node *tview.TreeNode) {
	//light := node.GetReference()
	//if light == nil {
	//return
	//} else {
	//node.SetExpanded(!node.IsExpanded())
	//}

	////children := node.GetChildren()
	////if len(children) == 0 {
	////// Toggle expanded
	////node.SetExpanded(!node.IsExpanded())
	////} else {
	////}
	//})
}
