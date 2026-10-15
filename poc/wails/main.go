package main

import (
	"embed"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// t0 is process start — the cold-start clock for the T-70 gate
// (startup ≤ 1.5 s to first frame, measured to the frontend's
// DOMContentLoaded → MarkFirstFrame round-trip).
var t0 = time.Now()

func main() {
	app := NewApp(t0)

	err := wails.Run(&options.App{
		Title:     "GitCurator — Wails PoC T-70",
		Width:     980,
		Height:    660,
		MinWidth:  980,
		MinHeight: 660,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// GitCurator LIGHT token "window" — warm cream background
		// (gui/theme.py:62); the whole point of the PoC: themed window.
		BackgroundColour: &options.RGBA{R: 251, G: 248, B: 242, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
