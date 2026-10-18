package connectioncheck

// githubsection.go — section 3: the GitHub token plus the vault
// backup repos (v0.83.0 — a missing repo is NEVER created
// automatically; 404 is a deliberate-setup hint, not a promise of
// auto-creation).

// CheckGitHub checks the token (GET /user) and every configured vault
// repo (GET /repos/{login}/{repo}): private → ok, public → warn
// (should be private!), 404 → the never-creates info line, anything
// else → warn. Never raises; a crashed check is the battery's job.
func CheckGitHub(config any, apiBase string, timeoutS float64) []Result {
	var out []Result
	token := trimPySpace(pyStr(pyOr(cfgValue(config, "github_token"), "")))
	if token == "" {
		out = append(out, result("GitHub token", LevelWarn,
			"not set — the app runs unauthenticated (60 req/h; "+
				"vault seals need a token)"))
		return out
	}
	r := GitHubGetFn("/user", token, apiBase, timeoutS)
	if r.Status == 0 {
		out = append(out, result("GitHub token", LevelError,
			"GitHub unreachable ("+r.Err+") — check internet / proxy"))
		return out
	}
	if r.Status == 401 {
		out = append(out, result("GitHub token", LevelError,
			"REJECTED (401 Bad credentials) — make a fresh token "+
				"at github.com/settings/tokens (classic, 'repo' scope)"))
		return out
	}
	if r.Status != 200 || !isDictLike(r.Payload) {
		out = append(out, result("GitHub token", LevelError,
			"unexpected response (HTTP "+itoa(r.Status)+")"))
		return out
	}
	login := pyStr(pyOr(dictValue(r.Payload, "login"), "?"))
	out = append(out, result("GitHub token", LevelOK,
		"valid — account "+login+" (5000 req/h)"))
	// The vault backup repos.
	seal := cfgValue(config, "vaultseal")
	type repoRow struct {
		label string
		repo  string
	}
	var repos []repoRow
	if pyTruthy(dictDefault(seal, "enabled", true)) {
		repo := trimPySpace(pyStr(pyOr(dictValue(seal, "repo_name"), "")))
		if repo != "" {
			repos = append(repos, repoRow{"Vault repo", repo})
		} else {
			out = append(out, result("Vault repo", LevelInfo,
				"not configured — no repo name set (Settings ▸ "+
					"Backup). Nothing is created or pushed until "+
					"you type one (remembered like credentials)"))
		}
	} else {
		out = append(out, result("Vault seal", LevelInfo,
			"disabled — the GitHub vault is not pushed (Settings ▸ Backup)"))
	}
	pipes := cfgValue(config, "pipelines")
	if pyTruthy(dictValue(pipes, "websites")) {
		repo := trimPySpace(pyStr(pyOr(cfgValue(config, "website_repo_name"), "")))
		if repo != "" {
			repos = append(repos, repoRow{"Websites repo", repo})
		}
	}
	// v0.84.0 — the Personal vault's backup repo joins the check (the
	// radar's third slot; no pipeline gate, exactly like its seal).
	repo := trimPySpace(pyStr(pyOr(cfgValue(config, "manual_repo_name"), "")))
	if repo != "" {
		repos = append(repos, repoRow{"Personal repo", repo})
	}
	for _, rr := range repos {
		r2 := GitHubGetFn("/repos/"+login+"/"+rr.repo, token, apiBase, timeoutS)
		switch {
		case r2.Status == 200 && isDictLike(r2.Payload):
			if pyTruthy(dictDefault(r2.Payload, "private", true)) {
				out = append(out, result(rr.label, LevelOK,
					rr.repo+" ready (private)"))
			} else {
				out = append(out, result(rr.label, LevelWarn,
					rr.repo+" is PUBLIC — vault backups should be "+
						"private (github.com/"+login+"/"+rr.repo+"/settings)"))
			}
		case r2.Status == 404:
			out = append(out, result(rr.label, LevelInfo,
				rr.repo+" not on GitHub yet — GitCurator never "+
					"creates repos automatically (v0.83.0): create "+
					"it yourself, or opt in once with "+
					"create_repo=true in config — the repo radar "+
					"(v0.84.0) finds and sets it on the next "+
					"start once it exists"))
		default:
			extra := ""
			if r2.Status == 0 && r2.Err != "" {
				extra = ", " + r2.Err
			}
			out = append(out, result(rr.label, LevelWarn,
				rr.repo+" could not be checked (HTTP "+itoa(r2.Status)+
					extra+")"))
		}
	}
	return out
}

// dictDefault is dict.get(key, default) on a nested dict value.
func dictDefault(m any, key string, def any) any {
	if v, ok := cfgGet(m, key); ok {
		return v
	}
	return def
}
