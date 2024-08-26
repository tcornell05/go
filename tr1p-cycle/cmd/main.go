package main

import (
	"flag"
	"io"
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/tcornell05/go/tr1p-cycle/internal/cycle"
)

var debug bool

func main() {
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.Parse()

	if debug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard) // Discard all logs when not in debug mode
	}

	myApp := app.New()

	defaultKeybinds := cycle.Keybinds{
		AddKeybind:    "Alt+Shift+E",
		RemoveKeybind: "Alt+Shift+D",
		CycleKeybind:  "Alt+Tab",
	}

	cl := cycle.NewCycleList()
	preview := cycle.NewPreview(myApp, cl)

	listener, err := cycle.NewKeybindListener(cl, defaultKeybinds, preview)
	if err != nil {
		log.Fatalf("Failed to create keybind listener: %v", err)
	}

	go cl.MonitorActiveWindow()
	go listener.Listen()

	myApp.Run()
}
