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

// SoundSuccessWav / SoundRetryWav are the batch-finish chimes
// (T-77 ships the files — the deferred-debt note's "until T-77/T-80
// ship the files" closes here). Synthesized assets (sine partials +
// exponential envelopes — see sounds/README.txt, no third-party
// material); the disk copy in APP_DIR/assets/sounds still WINS when
// present, so an owner can swap the chime by dropping a file (the
// embed is the fallback, never the override).
//
//go:embed sounds/success.wav
var SoundSuccessWav []byte

//go:embed sounds/retry.wav
var SoundRetryWav []byte
