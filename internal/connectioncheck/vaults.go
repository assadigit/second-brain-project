package connectioncheck

// vaults.go — section 1: the vault map as check rows (found AND
// writable — the repo's first writability probe). check_vault's kind
// ladder: 'github' (must exist), 'websites' (may be created by its
// pipeline), 'manual' (the owner creates it — the mirror never does).

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// VaultWritable writes + deletes a temp file inside path. Returns
// (ok, detail) — detail carries the exception on failure. Never
// raises, never leaves a file behind. The temp name rides the
// ".gc-write-test-" prefix (pinned by tests).
func VaultWritable(path string) (bool, string) {
	f, err := os.CreateTemp(path, ".gc-write-test-")
	if err != nil {
		return false, vaultErrDetail(err, "")
	}
	tmp := f.Name()
	if _, werr := f.Write([]byte("ok")); werr != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return false, vaultErrDetail(werr, tmp)
	}
	if cerr := f.Close(); cerr != nil {
		_ = os.Remove(tmp)
		return false, vaultErrDetail(cerr, tmp)
	}
	if rerr := os.Remove(tmp); rerr != nil {
		return false, vaultErrDetail(rerr, tmp)
	}
	return true, ""
}

// vaultErrDetail renders "{ExceptionClass}: {message}" — Python's
// OSError subclasses mapped where they exist (FileNotFoundError /
// PermissionError), helpers.PyOSErrorText for the message body.
func vaultErrDetail(err error, path string) string {
	return pyOSErrorClass(err) + ": " + pyOSErrorMessage(err, path)
}

func pyOSErrorClass(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return "FileNotFoundError"
		case errors.Is(err, fs.ErrPermission):
			return "PermissionError"
		}
	}
	s := errClass(err)
	return s
}

func pyOSErrorMessage(err error, path string) string {
	if path == "" {
		var pe *fs.PathError
		if errors.As(err, &pe) {
			path = pe.Path
		}
	}
	return helpers.PyOSErrorText(err, path)
}

// CheckVault is one vault row. kind: "github" (must exist), "websites"
// (may be created by its pipeline), "manual" (the owner creates it).
func CheckVault(path any, kind, label string) Result {
	p := trimPySpace(pyStr(pyOr(path, "")))
	if p == "" {
		switch kind {
		case "github":
			return result(label, LevelError,
				"not set — pick it in Settings ▸ Vault "+
					"(the app writes its notes here)")
		case "websites":
			return result(label, LevelWarn,
				"not set — the websites pipeline needs it "+
					"(Settings ▸ Vault)")
		default:
			return result(label, LevelInfo,
				"not set — the Library mirror is optional "+
					"(Settings ▸ Vault)")
		}
	}
	if !isDir(p) {
		if kind == "websites" {
			parent := filepath.Dir(rstripSlashBackslash(p))
			if parent == "" {
				parent = "."
			}
			if isDir(parent) {
				okW, why := VaultWritable(parent)
				if okW {
					return result(label, LevelInfo,
						"not on disk yet — created by the pipeline "+
							"on its first run ("+p+")")
				}
				return result(label, LevelError,
					"cannot be created — the parent folder is not "+
						"writable ("+why+")")
			}
			return result(label, LevelError,
				"cannot be created — parent folder missing ("+parent+")")
		}
		if kind == "manual" {
			return result(label, LevelWarn,
				"not found — create the vault in Obsidian yourself; "+
					"the mirror never creates it ("+p+")")
		}
		return result(label, LevelError, "missing on disk ("+p+")")
	}
	okW, why := VaultWritable(p)
	obs := ""
	if isDir(filepath.Join(p, ".obsidian")) {
		obs = " · Obsidian vault"
	}
	if okW {
		return result(label, LevelOK,
			"found · writable — ready to receive notes"+obs+" ("+p+")")
	}
	return result(label, LevelError,
		"found but NOT writable — the app cannot save notes here ("+why+")")
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// CheckVaults is the vault map as check rows: GitHub vault (always),
// Websites vault (checked when its pipeline is ON or a path is set),
// Manual vault.
func CheckVaults(config any) []Result {
	out := []Result{CheckVault(cfgValue(config, "vault_path"),
		"github", "GitHub vault")}
	pipes := cfgValue(config, "pipelines")
	wsOn := pyTruthy(dictValue(pipes, "websites"))
	wp := trimPySpace(pyStr(pyOr(cfgValue(config, "website_vault_path"), "")))
	if wsOn {
		out = append(out, CheckVault(wp, "websites", "Websites vault"))
	} else if wp != "" {
		out = append(out, result("Websites vault", LevelInfo,
			"set ("+wp+") · websites pipeline is OFF — not used "+
				"until you turn it on"))
	} else {
		out = append(out, result("Websites vault", LevelInfo,
			"not set · websites pipeline is OFF — nothing to check"))
	}
	mp := trimPySpace(pyStr(pyOr(cfgValue(config, "manual_vault_path"), "")))
	if mp != "" {
		out = append(out, CheckVault(mp, "manual", "Manual vault"))
	} else {
		out = append(out, result("Manual vault", LevelInfo,
			"not set — the Library mirror is optional"))
	}
	return out
}

// dictValue reads a key from a nested dict value (nil when absent).
func dictValue(m any, key string) any {
	v, _ := cfgGet(m, key)
	return v
}
