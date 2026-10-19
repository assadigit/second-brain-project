// Package fixture generates the PoC's synthetic vault — deterministic
// bytes, no owner data: N managed notes spread across real CATEGORY_FOLDERS
// trees (each with a source: line inside the head-800 window, the managed
// law) plus a few unmanaged stubs (no source line, never adopted).
//
// Scale: the reference's real vault is 1,309 notes (doc 07 §B.4's scan
// metric); the PoC's default 800-note fixture is the honest PoC-scale
// twin of that walk.
package fixture

import (
	"fmt"
	"os"
	"path/filepath"
)

// realFolders mirrors the reference CATEGORY_FOLDERS folder names (the
// constants.py:315-341 set — the same trees the engine's FolderCategory
// resolves). Duplicated deliberately: the webui module CANNOT import the
// main module's internal/constants (Go's internal-package rule —
// ADR-0002 Am.3's recorded boundary), and the fixture needs no engine
// code to lay out the tree.
var realFolders = []string{
	"AI-Domain/Agents",
	"AI-Domain/Agents/Frameworks",
	"AI-Domain/Agents/Implementations",
	"AI-Domain/Agents/Skills",
	"AI-Domain/MCP",
	"AI-Domain/Skills",
	"AI-Domain/LLM-Tools",
	"Tools/Scraping",
	"Tools/Automation",
	"Tools/Dev-Tools",
	"Tools/Networking",
	"Tools/Media",
	"Documentation/Guides",
	"Documentation/References",
	"Documentation/Research",
	"Frameworks/Web-Frameworks",
	"Frameworks/Backend",
	"Frameworks/Frontend",
	"Infrastructure",
	"Infrastructure/Deployment",
	"Infrastructure/Cloud",
	"Infrastructure/Containerization",
	"Infrastructure/Documentation",
	"Uncategorized",
}

// Vault describes the generated fixture.
type Vault struct {
	Root      string // the vault root (a caller-owned temp dir)
	Notes     int    // managed notes written
	Unmanaged int    // unmanaged stubs written
}

// WriteVault lays the fixture into root (created if missing). Deterministic:
// note i lands in realFolders[i%len(realFolders)] with content derived only
// from i — same bytes on every machine, every run.
func WriteVault(root string, notes int) (*Vault, error) {
	if notes < 0 {
		return nil, fmt.Errorf("fixture: notes must be >= 0, got %d", notes)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	for i := 0; i < notes; i++ {
		folder := filepath.Join(root, realFolders[i%len(realFolders)])
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return nil, err
		}
		name := fmt.Sprintf("fixture-note-%04d.md", i)
		body := fmt.Sprintf("---\ntitle: Fixture Note %04d\ntags: [t-70, fixture]\n---\n"+
			"source: https://github.com/fixture-org/repo-%04d\n\n"+
			"# Fixture Note %04d\n\nSynthetic body for the T-70 PoC baseline job.\n"+
			"Deterministic bytes, no owner data — line two of the body, index %d.\n",
			i, i, i, i)
		if err := os.WriteFile(filepath.Join(folder, name), []byte(body), 0o644); err != nil {
			return nil, err
		}
	}
	// a few unmanaged stubs (no source line — the never-adopted law).
	unmanaged := 2 + notes/200
	for j := 0; j < unmanaged; j++ {
		folder := filepath.Join(root, "Documentation/Guides")
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return nil, err
		}
		name := fmt.Sprintf("fixture-stub-%02d.md", j)
		body := fmt.Sprintf("---\ntitle: Fixture Stub %02d\n---\n\n# Stub %02d\n\nNo source line — unmanaged, never adopted by the baseline.\n", j, j)
		if err := os.WriteFile(filepath.Join(folder, name), []byte(body), 0o644); err != nil {
			return nil, err
		}
	}
	return &Vault{Root: root, Notes: notes, Unmanaged: unmanaged}, nil
}
