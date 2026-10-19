// Command webui is the T-70 PoC C GUI-shell gate (Phase 6's entry):
// the themed Wails window (the spec-gui-theme-icons LIGHT tokens), the
// coalesced log desk (deskbus + the frontend's cap/filter/tooltip laws),
// and ONE REAL JOB wired through — the note-state baseline job over a
// generated fixture vault, via the guibridge interim seam.
//
// It is a NESTED Go module (module github.com/assadigit/second-brain-
// project/webui) so the main module's runtime dependency set stays exactly
// the authorized four (ADR-0002 Am.3). The production shell (T-72)
// dissolves this module into the main tree.
package main

import (
	"embed"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

// t0 is process start — the cold-start clock for the T-70 gate
// (startup ≤ 1.5 s to first frame, measured to the frontend's
// DOMContentLoaded → MarkFirstFrame round-trip).
var t0 = time.Now()

func main() {
	app := NewApp(t0)

	err := wails.Run(&options.App{
		Title:     "GitCurator — T-70 GUI Shell Gate",
		Width:     1000,
		Height:    750,
		MinWidth:  760,
		MinHeight: 380,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// GitCurator LIGHT token "window" — the cream background
		// (gui/theme.py:58-237, spec-gui-theme-icons §1); the whole point
		// of the gate: the themed window.
		BackgroundColour: &options.RGBA{R: 251, G: 248, B: 242, A: 1},
		Windows: &windows.Options{
			// The production exe ships as a windowsgui-subsystem binary
			// (doc 04 §1); the PoC window carries the same no-console
			// intent on the Windows cross-build leg.
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
