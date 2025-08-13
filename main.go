package main

import (
	"context"
	"embed"

	"github.com/NeichS/final-redes-wails/internal/app"
	"github.com/NeichS/final-redes-wails/internal/client"
	sv "github.com/NeichS/final-redes-wails/internal/server"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var Assets embed.FS

func main() {
	// Create an instance of the app structure
	app := app.NewApp()
	server := &sv.FileServer{}
	client := &client.Receiver{}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "File transfer",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: Assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        func(ctx context.Context) {
			app.StartContext(ctx)
			server.StartContext(ctx)
		},
		Bind: []interface{}{
			app,
			server,
			client,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
