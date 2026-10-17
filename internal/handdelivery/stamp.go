// stamp.go — the master-table queue stamp (stamp_hand_rows,
// hand_delivery.py:776-830): rewrite the chosen rows' Status cells to
// the stamped hand gesture (🖐 hand — queued <date>) — the table's own
// feedback loop, so a hand-set cell is confirmed the same way
// dead/reviewed cells are.
package handdelivery

import (
	"fmt"
	"os"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// StampHandRows mirrors stamp_hand_rows: rewrite the chosen rows'
// Status cells to the stamped hand gesture (“ 🖐 hand — queued <date>
// “). URLs match by their RAW cell text (no canonical folding — the
// caller passes exactly the rows it means, unlike the harvest stamp's
// canonical bridge); any Status is stamped (the stronger verdicts are
// the CALLER's first checks — the consume table's pass 1.5 owns the
// precedence law). Pure file edit; never raises. Returns rows
// stamped.
func StampHandRows(vaultPath string, urls []string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	path := decomtable.DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return 0
	}
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	wanted := map[string]bool{}
	for _, u := range urls {
		wanted[u] = true
	}
	if len(wanted) == 0 {
		return 0
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log(fmt.Sprintf("⚠️ Hand stamp skipped — the table could not be read: %v", err), "warning")
		return 0
	}
	lines := pySplitLines(string(data))
	rows := decomtable.ParseDecommissionRows(lines)
	dateStr := dateNow()
	stamped := 0
	for _, row := range rows {
		if !wanted[row.URL] {
			continue
		}
		parts := strings.Split(row.Raw, "|")
		if len(parts) < 8 {
			continue
		}
		parts[6] = fmt.Sprintf(" 🖐 hand — queued %s ", dateStr)
		if row.Line < len(lines) {
			lines[row.Line] = strings.Join(parts, "|")
			stamped++
		}
	}
	if stamped == 0 {
		return 0
	}
	now := stampNow()
	for i, ln := range lines {
		if strings.HasPrefix(ln, "> Last updated:") {
			lines[i] = "> Last updated: " + now
		}
	}
	if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(lines, "\n"), "\n")+"\n"); err != nil {
		log(fmt.Sprintf("⚠️ Hand stamp could not be written: %v", err), "warning")
		return 0
	}
	return stamped
}
