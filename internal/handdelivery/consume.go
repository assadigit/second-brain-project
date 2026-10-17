// consume.go — the batch consume side (consume_delivered,
// hand_delivery.py:520-583): take the delivered pages, mark them
// consumed in queue.json (the FILE stays — it is the record), and
// hand them to the pipeline as real fetches. Includes the v0.53
// auto-page re-verification seam (chrome_tabs.page_is_real) and the
// empty-body law.
package handdelivery

import (
	"fmt"
	"os"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
)

// PageIsRealFn is the chrome_tabs.page_is_real seam (the v0.53
// auto-page re-verification: a file whose DOM carries a Chrome ERR_
// code, a crash phrase or a challenge grammar is a page a crashed or
// walled Chrome session left behind — never the site's content).
// The reference wraps the call in a try/except that degrades to
// (True, ”) when the verdict cannot be asked; the port keeps the
// same law with the seam unset (nil) — T-44 internal/chrometabs
// wires the real verdict and lands the fifth door's discard battery
// with it.
var PageIsRealFn func(pageHTML, href string) (ok bool, reason string)

// autoPageIsReal mirrors _auto_page_is_real's honest degradation: the
// reference asks chrome_tabs.page_is_real on the UTF-8-replaced text
// and answers (True, ”) when the verdict cannot be asked (the
// import failing, the call raising). The port consults the
// PageIsRealFn seam; an unset seam or a panic degrades the same way —
// a verdict that cannot be asked is not a wall.
func autoPageIsReal(u string, body []byte) (bool, string) {
	if PageIsRealFn == nil {
		return true, ""
	}
	return func() (ok bool, reason string) {
		defer func() {
			if r := recover(); r != nil {
				ok, reason = true, ""
			}
		}()
		text := helpers.DecodeUTF8Replace(body)
		return PageIsRealFn(text, u)
	}()
}

// ConsumedPage mirrors one consume_delivered() output item:
// {'url', 'wall', 'body', 'file'} — the pipeline feeds each one
// through the normal extract/classify/write path.
type ConsumedPage struct {
	URL  string
	Wall string
	Body []byte
	File string
}

// ConsumeDelivered mirrors consume_delivered: take the delivered
// pages — read each one, mark it consumed in queue.json (the FILE
// stays — it is the record). v0.53.0 — a page the APP delivered
// (door 'auto') is re-verified first: one that fails the fifth door's
// verdict is DISCARDED (the app's own file, removed) and its link
// keeps waiting; the owner's saved pages are trusted as always.
// Dry-run aware (nothing is marked, the pages are still read — a
// dry-run batch SHOWS the hand-delivery without consuming it). Never
// raises.
func ConsumeDelivered(vaultPath string, log LogFn) []ConsumedPage {
	if log == nil {
		log = noopLog
	}
	delivered := CollectDelivered(vaultPath)
	if len(delivered) == 0 {
		return nil
	}
	var out []ConsumedPage
	for _, item := range delivered {
		body, err := os.ReadFile(item.File)
		if err != nil {
			log(fmt.Sprintf("⚠️ Hand-delivered page unreadable (%s): %v — the link keeps waiting", item.Suggested, err), "warning")
			continue
		}
		if pyBytesBlank(body) {
			log(fmt.Sprintf("⚠️ Hand-delivered page is empty (%s) — the link keeps waiting", item.Suggested), "warning")
			continue
		}
		if item.Door == "auto" {
			if ok, reason := autoPageIsReal(item.URL, body); !ok {
				discardAutoPage(item.File)
				log(fmt.Sprintf("⚠️ The auto-delivered page for %s was NOT the site (%s) — discarded (the app's own file), the link keeps waiting for its doors", item.URL, reason), "warning")
				continue
			}
		}
		out = append(out, ConsumedPage{
			URL: item.URL, Wall: item.Wall, Body: body, File: item.File,
		})
	}
	if len(out) == 0 {
		return nil
	}
	if !dryrun.IsEnabled() {
		queue := ReadQueue(vaultPath)
		for _, item := range out {
			if meta := queue.Meta(item.URL); meta != nil {
				meta.Set("consumed", stampNow())
			}
		}
		if err := queue.WriteAtomic(vaultPath); err != nil {
			log(fmt.Sprintf("⚠️ Hand-delivery queue could not be stamped: %v", err), "warning")
		}
	}
	log(fmt.Sprintf("🖐 Hand-delivery: %d delivered page(s) taken — processing them as real fetches", len(out)), "info")
	return out
}

// discardAutoPage mirrors _discard_auto_page: remove a page the app
// itself delivered that failed its own verdict — the file is OURS
// (the fifth door wrote it), never the owner's save, and leaving it
// would re-fail every consume pass. Best-effort; never raises.
func discardAutoPage(path string) {
	defer func() { recover() }()
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		_ = os.Remove(path)
	}
}

// pyBytesBlank mirrors Python's `not body.strip()` for bytes: true
// when every byte is ASCII whitespace (space, \t, \n, \r, \v, \f) —
// bytes.strip never strips non-ASCII (a NBSP-only body is NOT blank
// for the reference; the earlier strings.TrimSpace check was the
// latent divergence, reality-wins).
func pyBytesBlank(b []byte) bool {
	for _, c := range b {
		switch c {
		case ' ', '\t', '\n', '\r', '\v', '\f':
		default:
			return false
		}
	}
	return true
}
