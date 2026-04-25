package main

import (
	"embed"
	"log"

	"stargrazer/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := config.Init(); err != nil {
		log.Printf("config warning: %v (using defaults)", err)
	}

	win := config.GetWindow()
	app := NewApp()

	err := wails.Run(&options.App{
		Title:    win.Title,
		Width:    win.Width,
		Height:   win.Height,
		MinWidth: win.MinWidth,
		MinHeight: win.MinHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
