# T-78 — the side-by-side verification vs the Python app — evidence

Recorded 2026-10-20 (Asia/Tehran), SESSION-041. Environment: Debian 13 trixie sandbox, go1.27.2,
the stdlib-only venv (Python 3.12.15), the reference @ 98f4fc71cfb8 (PINNED-GITCURATOR-REV).

The machine: `testdiff/side_by_side.py` (committed) — Leg A drives the REFERENCE'S OWN
lock code, extracted verbatim by pinned line range from the read-only checkout
(gui/app.py:393-440 + gui/headless.py:246-282 — the PyQt6 import wall makes importing
the module impossible under the stdlib-only venv; the `--cli`/`--headless` entries
both exit BEFORE the lock dance, so the GUI entry is the reference's only
lock-taking binary), and drives the Go production door (internal/app.AcquireAppLock)
via `testdiff/goside`. Leg B runs the reference's own pure config modules
(storage/user_config/constants) in-process against goside's config/fill/mirror doors.

THE DIVERGENCE THE AUDIT CAUGHT AND FIXED: the Go twin's POSIX liveness probe read
EPERM (a live process owned by another user) as ALIVE; the reference's blanket
`except (OSError, …): return False` swallows the PermissionError and reads DEAD —
so a lock file whose stale PID was recycled to a stranger's process was take-over-able
in Python but refused by Go. Fixed in internal/app/process_posix.go (killVerdict,
pinned hermetically by TestKillVerdictMirrorsTheBlanketExcept); A6 proves BOTH
binaries now take such a lock over identically, live.

# T-78 — the side-by-side verification vs the Python app

- Verdict: **GREEN** (21/21 checks)
- Reference: /home/z/my-project/GitCurator/app @ 98f4fc71cfb8
- Python: /home/z/venvs/gitcurator/bin/python
- The lock extraction provenance: gui/app.py:393-440 sha256 8025651fcb52eb5c… · gui/headless.py:246-282 sha256 c96105db365077cb…

## Checks

- [PASS] A0 go fresh-take bytes
  - lock bytes = b'18687' (5 bytes), want the bare decimal b'18687'
  - hex: 3138363837
- [PASS] A0 py fresh-take bytes
  - lock bytes = b'18695' (5 bytes), want the bare decimal b'18695'
  - hex: 3138363935
- [PASS] A1 go-holds / python-refuses
  - exit=1 (want 1)
  - stdout: 'Another instance is already running (PID 18696). Exiting.'
  - the lock file still carries the Go holder's pid — untouched
- [PASS] A2 python-holds / go-refuses
  - exit=1 (want 1)
  - stdout: 'Another instance is already running (PID 18703). Exiting.\nGOSIDE REFUSED'
  - the lock file still carries the Python holder's pid — untouched
- [PASS] A3 go-crash / python-heals-stale
  - the stale pid was the crashed Go holder (18709)
  - stdout: 'Removing stale lock file (PID 18709 is no longer running)...\nPYLOCK HELD 18717'
- [PASS] A4 python-crash / go-heals-stale
  - the stale pid was the crashed Python holder (18718)
  - holder output: 'Removing stale lock file (PID 18718 is no longer running)...\nGOSIDE HELD 18719'
- [PASS] A4b the healed lock's bytes (go)
  - lock bytes = b'18719' (5 bytes), want the bare decimal b'18719'
  - hex: 3138373139
- [PASS] A6 eperm-probe py reads dead (take-over)
  - the probe pid 1 is alive but not ours to signal
  - stdout: 'Removing stale lock file (PID 1 is no longer running)...\nPYLOCK HELD 18726'
  - both binaries take the lock over — the blanket-except law
- [PASS] A6 eperm-probe go reads dead (take-over)
  - the probe pid 1 is alive but not ours to signal
  - stdout: 'Removing stale lock file (PID 1 is no longer running)...\nGOSIDE HELD 18727'
  - both binaries take the lock over — the blanket-except law
- [PASS] B1 both apps write the SAME bytes (merge + write laws)
  - sha256(go) = e15c880650e0f07f7450cd407e1f510230ad309051a756175d9aebd75af7ef61
  - sha256(py) = e15c880650e0f07f7450cd407e1f510230ad309051a756175d9aebd75af7ef61
  - goside: 'GOSIDE CONFIG fresh=false filled=0 mirrored=true' exit=0
- [PASS] B1a unknown keys survive both merges
  - the fixture's two unknown keys are in the Go-written bytes
- [PASS] B1b tuning keys unstomped (default-only)
  - the custom 99 / 7 / 2.5 values survive — never the 60/3/0.5 defaults
- [PASS] B1c nested dicts merge key-by-key
  - proxy.host edited + ollama.model added — the rest of each dict kept
- [PASS] B1d non-ASCII raw + no HTML escaping
  - the unicode+HTML-ish value's raw UTF-8 (37 bytes) is byte-present — ensure_ascii=False and SetEscapeHTML(false) both live
- [PASS] B1e python's re-save of the go file is byte-stable
  - sha256(py-re-save) = e15c880650e0f07f7450cd407e1f510230ad309051a756175d9aebd75af7ef61
- [PASS] B1f go's re-save of the python file is byte-stable
  - sha256(go-re-save) = e15c880650e0f07f7450cd407e1f510230ad309051a756175d9aebd75af7ef61
  - goside: 'GOSIDE CONFIG fresh=false filled=0 mirrored=true'
- [PASS] B2 fill (kept) — same bytes, same count
  - sha256(py) = bb43a7104ef3c47d9a13d086b4d0e0c65fadb71a296496a49d9077398fa4ad38
  - sha256(go) = bb43a7104ef3c47d9a13d086b4d0e0c65fadb71a296496a49d9077398fa4ad38
  - counts: py=3, go line='GOSIDE FILL 3'
- [PASS] B2a local-wins law (kept mode)
  - the real local value beats the remembered one
- [PASS] B2 fill (fresh) — same bytes, same count
  - sha256(py) = 16c8fd846e8cbb4fc682aab33e53c20935de2e6444e728027a60db0f80ffd047
  - sha256(go) = 16c8fd846e8cbb4fc682aab33e53c20935de2e6444e728027a60db0f80ffd047
  - counts: py=8, go line='GOSIDE FILL 8'
- [PASS] B3 mirror — same memory bytes, never-loses
  - sha256(py) = 248accd751ad0590b311d74d52beab5287a3fd9ec14322944fc3993832c6e92e  mode=0o600
  - sha256(go) = 248accd751ad0590b311d74d52beab5287a3fd9ec14322944fc3993832c6e92e  mode=0o600
  - goside: 'GOSIDE MIRROR true'
  - bot_token (memory-only) kept: True
- [PASS] B3a the 0600 permission law
  - py=0o600 go=0o600
