// cli.go — the standalone CLI (directory_sync.py's main, lines
// 1228-1292): `python -m gitcurator.cloud.directory_sync` — build the
// index, print/write it, the full --sync-only cycle, --publish, and
// --make-admin-hash (the Phase-2 owner setup helper's PBKDF2 hash).
package directorysync

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// Main mirrors main(argv): the standalone CLI's surface — --vault,
// --config (default ./config.json), --sync-only, --publish,
// --print-index, --out FILE, --make-admin-hash. The flag package's
// parse-error exit 2 stays the reference's argparse surface.
//
// The getpass convergence: Python's getpass hides the password on a
// TTY but itself falls back to a plain stdin read (with a warning)
// where it cannot; Go's stdlib has no cross-platform getpass, so the
// password is read plainly from stdin — the reference's own fallback
// door, documented in the spec's Go notes.
func Main(argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("directory_sync", flag.ContinueOnError)
	fs.SetOutput(stderr)
	vaultFlag := fs.String("vault", "",
		"Websites vault root (overrides config website_vault_path)")
	configFlag := fs.String("config", "config.json",
		"config.json path (default: ./config.json)")
	syncOnly := fs.Bool("sync-only", false,
		"the scheduled mode: apply pending admin commands → mirror → "+
			"publish → ack (the golden-rule-5 order, SPEC §10.3)")
	publish := fs.Bool("publish", false,
		"publish the built index (implied by --sync-only)")
	printIndex := fs.Bool("print-index", false,
		"print the built index JSON instead of publishing")
	makeAdminHash := fs.Bool("make-admin-hash", false,
		"ask for a password and print its PBKDF2 salted hash for "+
			"`wrangler secret put ADMIN_PASSWORD_HASH` (Phase 2 owner setup — "+
			"the password itself is never stored anywhere)")
	out := fs.String("out", "", "write the built index JSON to this file")
	if err := fs.Parse(argv); err != nil {
		return 2
	}

	if *makeAdminHash {
		return makeAdminHashCmd(stdin, stdout)
	}

	config := loadConfigFile(*configFlag)
	if *vaultFlag != "" {
		config.Set("website_vault_path", *vaultFlag)
	}
	cfg := effectiveConfig(config)
	vault := pyStrOrEmpty(mustCfgGet(cfg, "website_vault_path"))

	if vault == "" || !isDir(vault) {
		fmt.Fprintln(stderr, "[directory] vault not found — pass --vault or set "+
			"website_vault_path in config.json")
		return 2
	}

	log := func(msg string) { fmt.Fprintln(stdout, msg) }

	EnsureSystemFiles(vault, log)
	index := BuildIndex(vault, nil, log)

	if *printIndex {
		fmt.Fprintln(stdout, storage.PyDumps(index))
	}
	if *out != "" {
		if err := os.WriteFile(*out, []byte(storage.PyDumps(index)), 0o644); err == nil {
			fmt.Fprintln(stdout, "[directory] index written to "+*out)
		}
	}
	if *syncOnly {
		// v0.76.0 — --sync-only is the FULL Phase-2 cycle (SPEC §10.3):
		// commands first, then mirror, then publish, then ack.
		config.Set("website_vault_path", vault)
		summary := RunCommandCycle(config, log)
		fmt.Fprintln(stdout, "[directory] sync-only cycle: "+pyReprValue(summary))
		published := pyTruthyDS(mustCfgGet(summary, "published"))
		failed := int(pyIntOf(mustCfgGet(summary, "failed")))
		if published && failed == 0 {
			return 0
		}
		return 1
	}
	if *publish {
		ok := PublishIndex(index,
			pyStrOrEmpty(mustCfgGet(cfg, "directory_worker_url")),
			pyStrOrEmpty(mustCfgGet(cfg, "directory_sync_token")), 0, log)
		if ok != nil {
			return 0
		}
		return 1
	}
	if !*printIndex && *out == "" {
		counts, _ := dictDoorOf(index).Get("counts").(*storage.OrderedMap)
		sites := 0
		if counts != nil {
			sites = int(pyIntOf(mustCfgGet(counts, "sites")))
		}
		fmt.Fprintln(stdout, "[directory] built "+strconv.Itoa(sites)+
			" sites (pass --sync-only to publish)")
	}
	return 0
}

// loadConfigFile mirrors _load_config_file: json.load → dict, {} on
// every miss.
func loadConfigFile(path string) *storage.OrderedMap {
	data, err := os.ReadFile(path)
	if err != nil {
		return storage.NewOrderedMap()
	}
	if cfg, err := storage.DecodeJSON(data); err == nil {
		return cfg
	}
	return storage.NewOrderedMap()
}

// makeAdminHashCmd mirrors _make_admin_hash: ask for a password
// (twice), print its PBKDF2 salted hash and the exact wrangler
// command. The password itself is never stored, printed, or logged —
// only the hash leaves this function.
func makeAdminHashCmd(stdin io.Reader, stdout io.Writer) int {
	reader := bufio.NewReader(stdin)
	readLine := func() (string, bool) {
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return "", false
		}
		return strings.TrimRight(line, "\r\n"), true
	}
	password, ok1 := readLine()
	confirm, ok2 := readLine()
	if !ok1 || !ok2 {
		fmt.Fprintln(stdout, "[admin] could not read the password: EOF")
		return 2
	}
	if password == "" || len(password) < 8 {
		fmt.Fprintln(stdout, "[admin] the password must be at least 8 characters")
		return 2
	}
	if password != confirm {
		fmt.Fprintln(stdout, "[admin] the two entries do not match — nothing was done")
		return 2
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		fmt.Fprintln(stdout, "[admin] could not read the password: "+err.Error())
		return 2
	}
	iterations := 100000
	derived := pbkdf2HMACSHA256([]byte(password), salt, iterations, 32)
	spec := "pbkdf2$" + strconv.Itoa(iterations) + "$" + hex.EncodeToString(salt) +
		"$" + hex.EncodeToString(derived)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Set it on the Worker (app/cloudflare-bot folder):")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "  npx wrangler secret put ADMIN_PASSWORD_HASH")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "and paste this line when it asks:")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "  "+spec)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "(A salted PBKDF2 hash — safe to paste, useless to a thief.")
	fmt.Fprintln(stdout, " Forgot the password? Just run this again.)")
	return 0
}

// pbkdf2HMACSHA256 is RFC 2898's PBKDF2 with HMAC-SHA256 —
// hashlib.pbkdf2_hmac('sha256', …) hand-rolled (stdlib only; the
// x/crypto/pbkdf2 door stays shut per the dependency law).
func pbkdf2HMACSHA256(password, salt []byte, iterations, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	numBlocks := (keyLen + prf.Size() - 1) / prf.Size()
	var dk []byte
	buf := make([]byte, 4)
	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(buf, uint32(block))
		prf.Write(buf)
		u := prf.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)
		for i := 1; i < iterations; i++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}
