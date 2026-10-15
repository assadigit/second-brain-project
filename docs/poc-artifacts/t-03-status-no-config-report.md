# Differential report — status-no-config

- Run: 2026-10-15 04:55:48Z (UTC)
- Reference: 98f4fc71cfb862cee7225d864636c08062724348 (pinned: 98f4fc71cfb862cee7225d864636c08062724348)
- Go binary: /home/z/second-brain-project/testdiff/runs/bin/gitcurator-cli (sha256 51590284566c907e…)
- Python: /home/z/venvs/gitcurator/bin/python
- Scratch cwd: `/home/z/second-brain-project/testdiff/runs/status-no-config/scratch` (normalized to `<TMP>` where it appears)
- Normalization rules applied: temp-path, timestamp

## Commands (same scratch cwd, same env)

```
python: /home/z/venvs/gitcurator/bin/python /home/z/GitCurator/app/main.py --cli --status --config missing-config.json
go:     /home/z/second-brain-project/testdiff/runs/bin/gitcurator-cli --status --config missing-config.json
```

## Raw outputs (sha256)

- python stdout: `7f118c3cfa88f44c2eeb054522b9935785aab0820cfdadf3554dfc10c4fd533f` (810 bytes)
- go stdout:     `7f118c3cfa88f44c2eeb054522b9935785aab0820cfdadf3554dfc10c4fd533f` (810 bytes)
- python stderr: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` (0 bytes)
- go stderr:     `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` (0 bytes)

## Verdict

**PASS — outputs identical after normalization; exit codes match**

Normalized stdout (python side shown; go side byte-identical after the rules above):

```

   ▄███████▄  ▄████████    ▄████████    ▄████████  ███▄▄▄▄      ▄████████
  ███    ███ ███    ███   ███    ███   ███    ███ ███▀▀▀██▄   ███    ███
  ███    ███ ███    ███   ███    ███   ███    █▀▀ ███    ███  ███    ███
 ▄███▄▄▄▄███▄███▄▄▄▄▄███▄ ███    ███  ▄███▄▄▄     ███    ███ ▄███▄▄▄▄███▄
▀▀▀▀▀▀▀▀▀▀▀ ███▀▀▀▀▀▀▀▀▀  ███    ███ ▀▀▀▀▀▀▀    ▀▀     ███▀ ▀▀▀▀▀   ███▀
▀███████████ ███    ███   ███    ███ ▀███████████ ▄██    ███ ▄██████████▄
  ███    ███ ███    ███   ███    ███          ███ ███    ███ ███    █████
  ███    ███ ███    ███ ████    ████ ▄███████████  ▀█████▀  ███    █████
  GitHub Project Curator — CLI v0.09.5  ·  colored · animated · fully automatic

<TS> ▲ No config at <TMP>/missing-config.json — run `--cli --init` first.

```

Normalized stderr (both sides):

```
(empty)
```