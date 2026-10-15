// pylower.go — the T-17 promotion: the implementation of Python's
// str.lower() MOVED to internal/helpers (the canonical home, corpus
// included); this delegation keeps links' call sites and its golden
// tests stable.
package links

import (
	"github.com/assadigit/second-brain-project/internal/helpers"
)

// pyLower mirrors Python's str.lower() (see helpers.PyLower).
func pyLower(s string) string { return helpers.PyLower(s) }
