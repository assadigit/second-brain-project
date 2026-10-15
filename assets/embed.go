// Package assets embeds the application's static resources (prompts,
// taxonomy defaults, sounds, fonts, icons) per
// docs/port-blueprint/04-target-go-architecture.md §1.
package assets

import _ "embed"

// IconICO is the ready-made brand mark (gitcurator.ico) used by the Wails
// window and the Windows executable (owner decision: use the repo's brand
// mark, swap the file later if a custom one arrives — ADR-0010 Am.2).
//
//go:embed icons/gitcurator.ico
var IconICO []byte
