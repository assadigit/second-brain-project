// Command webui is the production Wails shell (T-72): the themed
// window over the production composition — internal/app bound DIRECTLY
// (the guibridge/ interim seam retired at this landing; the nested
// webui module may import the parent's internal/ packages, ADR-0002
// Am.4's probe). The window's colors ride the golden-locked token
// tables (webui/uitheme); the desk rides the deskbus + the cache-mirror
// frontend half (THE COALESCED LOG contract).
//
// It is still a NESTED Go module (module github.com/assadigit/second-
// brain-project/webui): the T-72 dissolution probe (ADR-0002 Am.5)
// found wails v2.16.0's own go.mod demands x/net v0.56.0 + the 1.25.0
// directive — bumps the authorized runtime set forbids — so the
// boundary stays nested and the direct binding rides Am.4's probe.
package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/assadigit/second-brain-project/webui/uitheme"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

// t0 is process start — the cold-start clock (startup ≤ 1.5 s to the
// first frame, measured to the frontend's DOMContentLoaded →
// MarkFirstFrame round-trip).
var t0 = time.Now()

// version is the release stamp (doc 10 §5): "dev" in a bare build;
// the release ritual stamps it at compile time —
//
//	wails build -platform windows/amd64 -trimpath \
//	    -ldflags "-X main.version=$(cat ../VERSION)"
//
// The CLI twin carries the same var (cmd/gitcurator-cli/main.go — the
// --version door prints it). The stamp's byte-presence inside BOTH
// exes is asserted by build/mkzip's version-stamp guard (a forgot-the-
// ldflags build never reaches the zip); the wails.json Info block's
// productVersion is pinned to the repo VERSION file by test (the
// one-version-truth law).
var version = "dev"

func main() {
	// The reference's own boot stamp (gui/app.py:30-38): "[main]
	// LOADED version …" printed to stderr at GUI startup "so you
	// can verify the right file loads" — the Python blob is the
	// legacy __VERSION__ string; the Go twin speaks the release
	// stamp. This print is also what keeps the -X value alive:
	// the linker drops an UNREFERENCED var's -X bytes under
	// -trimpath (T-80's empirical find), and the release zip's
	// version-stamp guard reads those bytes. The reference skips
	// the print in --cli mode (the blob broke CLI output); the Go
	// GUI has no CLI mode, so the print is unconditional — and
	// invisible on the windowsgui subsystem (no console).
	fmt.Fprintf(os.Stderr, "[main] LOADED version %s\n", version)

	shell := NewShell(t0)

	r, g, b, a := uitheme.RGBA(uitheme.Light, "window")
	err := wails.Run(&options.App{
		Title: "GitCurator",
		// THE CONTAINER LAW (the owner's SESSION-075 directive —
		// "the app container is too big it must be like 20% smaller":
		// 1000x750 -> 800x600, exactly -20%; the mins stay (760x380 —
		// the resize floor the layout keeps honoring).
		Width:     800,
		Height:    600,
		MinWidth:  760,
		MinHeight: 380,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// The LIGHT token table's "window" — the cream background
		// (gui/theme.py:58-237, golden-locked); the dark theme's
		// window is the frontend's [data-theme] flip (T-77's toggle).
		BackgroundColour: &options.RGBA{R: r, G: g, B: b, A: a},
		Windows: &windows.Options{
			// The production exe ships as a windowsgui-subsystem binary
			// (doc 04 §1); the no-console intent carries to the
			// Windows cross-build leg.
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		OnStartup:  shell.startup,
		OnShutdown: shell.shutdown,
		Bind: []interface{}{
			shell,
			shell.settingsUI,
			shell.pagesBUI,
			shell.syncUI,
			shell.scanUI,
			shell.heroUI,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
