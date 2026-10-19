// goside — the Go half of T-78's side-by-side differential
// (testdiff/side_by_side.py drives both halves against the SAME data).
//
// This is test instrumentation, not a shipped command: it exposes the
// production doors (internal/app's AcquireAppLock — the exact lock the
// wails shell takes before its window; internal/config's app-door
// load/save sequence — LoadAppConfig + storage.MergeConfig +
// storage.WriteConfigFile + MirrorCredentialsToGlobal, the same doors
// the shells call) behind mode verbs a harness can spawn:
//
//	goside lock-hold <app.lock>    take the lock, speak the verdict,
//	                              HOLD until SIGTERM/SIGINT (release on
//	                              the way out). Exit 1 when refused —
//	                              the shells' law (the alive-PID refusal
//	                              exits 1); 0 otherwise.
//	goside lock-try <app.lock>     one-shot take-or-refuse: speak the
//	                              verdict, release, exit 1 when refused.
//	goside config-save <config.json> <updates.json>
//	                              the app door's save sequence: the
//	                              load (fresh detect + example base +
//	                              THE FILL + the strip heal), the merge
//	                              (merge-never-rebuild), the atomic
//	                              byte-comparable write, THE MIRROR.
//	                              Prints the machine verdict line.
//	goside fill <.gitcurator> <cfg-in> <cfg-out> <fresh>
//	                              ApplyGlobalCredentials in isolation:
//	                              the tree from cfg-in, the fill from the
//	                              memory file, the result to cfg-out.
//	goside mirror <.gitcurator> <cfg-in>
//	                              MirrorCredentialsToGlobal in isolation.
//
// The user-memory path flows through GlobalConfigPath()'s env override
// (GITCURATOR_USER_CONFIG) exactly like the shells — the harness steers
// both sides at the same scratch memory file.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/config"
	"github.com/assadigit/second-brain-project/internal/storage"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: goside <lock-hold|lock-try|config-save|fill|mirror> ... (see the file header)")
	os.Exit(2)
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}
	switch os.Args[1] {
	case "lock-hold", "lock-try":
		lockMode(os.Args[1], os.Args[2])
	case "config-save":
		if len(os.Args) != 4 {
			usage()
		}
		configSave(os.Args[2], os.Args[3])
	case "fill":
		if len(os.Args) != 6 {
			usage()
		}
		fresh, err := strconv.ParseBool(os.Args[5])
		if err != nil {
			usage()
		}
		fillMode(os.Args[2], os.Args[3], os.Args[4], fresh)
	case "mirror":
		if len(os.Args) != 4 {
			usage()
		}
		mirrorMode(os.Args[2], os.Args[3])
	default:
		usage()
	}
}

// lockMode runs the production lock door. AcquireAppLock speaks the
// reference's own console lines; the machine line afterwards is the
// harness's parse target.
func lockMode(mode, path string) {
	l := app.AcquireAppLock(path)
	switch {
	case l.Refused:
		fmt.Printf("GOSIDE REFUSED\n")
		os.Exit(1)
	case !l.Held:
		// The reference's belt: any other failure CONTINUES WITHOUT the
		// lock (app.py's "Cannot create lock file" path).
		fmt.Printf("GOSIDE NOLOCK\n")
		os.Exit(0)
	}
	fmt.Printf("GOSIDE HELD %d\n", os.Getpid())
	if mode == "lock-try" {
		l.Release()
		return
	}
	// Hold until a clean stop; the harness's kill -9 (the stale-lock
	// scenarios) skips this and leaves the file behind — by design.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
	<-ch
	l.Release()
}

// configSave is the app door's save sequence over the given config.json.
func configSave(cfgPath, updatesPath string) {
	res := config.LoadAppConfig(cfgPath)
	updates := readTree(updatesPath)
	merged := storage.MergeConfig(res.Config, updates)
	if err := storage.WriteConfigFile(cfgPath, merged); err != nil {
		fmt.Fprintf(os.Stderr, "GOSIDE WRITE-FAILED %v\n", err)
		os.Exit(1)
	}
	mirrored := config.MirrorCredentialsToGlobal(merged, "")
	fmt.Printf("GOSIDE CONFIG fresh=%t filled=%d mirrored=%t\n",
		res.Fresh, res.Filled, mirrored)
}

func fillMode(memPath, inPath, outPath string, fresh bool) {
	tree := readTree(inPath)
	n := config.ApplyGlobalCredentials(tree, memPath, fresh)
	if err := storage.WriteConfigFile(outPath, tree); err != nil {
		fmt.Fprintf(os.Stderr, "GOSIDE WRITE-FAILED %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("GOSIDE FILL %d\n", n)
}

func mirrorMode(memPath, inPath string) {
	tree := readTree(inPath)
	ok := config.MirrorCredentialsToGlobal(tree, memPath)
	fmt.Printf("GOSIDE MIRROR %t\n", ok)
}

// readTree decodes a JSON file as the ordered tree ({} when absent —
// the harness always provides its inputs; an absent updates file is the
// empty merge).
func readTree(path string) *storage.OrderedMap {
	raw, err := os.ReadFile(path)
	if err != nil {
		return storage.NewOrderedMap()
	}
	tree, err := storage.DecodeJSON(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GOSIDE BAD-JSON %s: %v\n", path, err)
		os.Exit(1)
	}
	return tree
}
