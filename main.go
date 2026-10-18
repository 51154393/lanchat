package main

import (
	"embed"
	"lanchat/internal/app"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	application := app.NewApp()

	err := wails.Run(&options.App{
		Title:     "LAN Chat",
		Width:     1024,
		Height:    600,
		MinWidth:  800,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 237, G: 237, B: 237, A: 1},
		OnStartup:        application.Startup,
		OnShutdown:       application.Shutdown,
		Bind: []interface{}{
			application,
		},
		StartHidden: false,
		Windows: &windows.Options{
			Theme:        windows.Light,
			BackdropType: windows.Mica,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:           windows.RGB(252, 252, 252),
				DarkModeTitleBarInactive:   windows.RGB(252, 252, 252),
				DarkModeTitleText:          windows.RGB(25, 25, 25),
				DarkModeTitleTextInactive:  windows.RGB(160, 160, 160),
				DarkModeBorder:             windows.RGB(231, 231, 231),
				DarkModeBorderInactive:     windows.RGB(240, 240, 240),
				LightModeTitleBar:          windows.RGB(252, 252, 252),
				LightModeTitleBarInactive:  windows.RGB(252, 252, 252),
				LightModeTitleText:         windows.RGB(25, 25, 25),
				LightModeTitleTextInactive: windows.RGB(160, 160, 160),
				LightModeBorder:            windows.RGB(231, 231, 231),
				LightModeBorderInactive:    windows.RGB(240, 240, 240),
			},
		},
	})

	if err != nil {
		log.Printf("Error: %v", err)
	}
}
