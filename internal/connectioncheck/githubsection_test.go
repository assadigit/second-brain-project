package connectioncheck

// githubsection_test.go — section 3: the token + the vault repo
// ladder against a fake api.github.com (the 401/404/public-repo laws,
// the v0.83 never-creates line, the seal gates) + THE UA TABLE's
// _github_get entry (Bearer + Accept + the connection UA). Ported from
// the reference's TestCheckGithub.

import (
	"strings"
	"testing"
)

func githubCfg(kw map[string]any) map[string]any {
	base := map[string]any{
		"github_token": "tok",
		"vaultseal":    map[string]any{"enabled": true, "repo_name": "backup"},
		"pipelines":    map[string]any{"websites": false},
	}
	for k, v := range kw {
		base[k] = v
	}
	return base
}

func TestCheckGithubNoTokenIsWarnSingleRow(t *testing.T) {
	rows := CheckGitHub(map[string]any{"github_token": ""},
		GitHubAPIBase, 15.0)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].Level != LevelWarn {
		t.Errorf("level = %s, want warn", rows[0].Level)
	}
	if !strings.Contains(rows[0].Detail, "not set") {
		t.Errorf("detail: %q", rows[0].Detail)
	}
}

func TestCheckGithubValidTokenPrivateRepoReady(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user":                  route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup": route(200, map[string]any{"private": true}),
	})
	rows := CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	for i, want := range []string{LevelOK, LevelOK} {
		if rows[i].Level != want {
			t.Errorf("row %d level = %s, want %s", i, rows[i].Level, want)
		}
	}
	if !strings.Contains(rows[0].Detail, "account testuser") {
		t.Errorf("detail: %q", rows[0].Detail)
	}
	if !strings.Contains(rows[0].Detail, "5000 req/h") {
		t.Errorf("rate context missing: %q", rows[0].Detail)
	}
	if !strings.Contains(rows[1].Detail, "backup ready (private)") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
}

func TestCheckGithubPublicRepoShouldBePrivateWarn(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user":                  route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup": route(200, map[string]any{"private": false}),
	})
	rows := CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	if rows[1].Level != LevelWarn {
		t.Fatalf("level = %s, want warn", rows[1].Level)
	}
	if !strings.Contains(rows[1].Detail, "PUBLIC") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
}

func TestCheckGithubMissingRepoNeverCreatedAutomaticallyV0830(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user":                  route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup": route(404, map[string]any{"message": "Not Found"}),
	})
	rows := CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	if rows[1].Level != LevelInfo {
		t.Fatalf("level = %s, want info", rows[1].Level)
	}
	// v0.83.0 — the auto-creation promise is dead: 404 now says the
	// owner creates it (or opts in once with create_repo=true).
	if !strings.Contains(rows[1].Detail, "never creates repos automatically") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
	if !strings.Contains(rows[1].Detail, "create_repo=true") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
}

func TestCheckGithubRejectedToken401(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user": route(401, map[string]any{"message": "Bad credentials"}),
	})
	rows := CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].Level != LevelError {
		t.Errorf("level = %s, want error", rows[0].Level)
	}
	if !strings.Contains(rows[0].Detail, "REJECTED") {
		t.Errorf("detail: %q", rows[0].Detail)
	}
}

func TestCheckGithubUnexpectedResponseIsError(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user": route(500, map[string]any{"message": "boom"}),
	})
	rows := CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	if len(rows) != 1 || rows[0].Level != LevelError {
		t.Fatalf("rows = %+v", rows)
	}
	if !strings.Contains(rows[0].Detail, "unexpected response (HTTP 500)") {
		t.Errorf("detail: %q", rows[0].Detail)
	}
}

func TestCheckGithubUnreachable(t *testing.T) {
	rows := CheckGitHub(githubCfg(nil), deadURL(t), 2.0)
	if rows[0].Level != LevelError {
		t.Fatalf("level = %s", rows[0].Level)
	}
	if !strings.Contains(rows[0].Detail, "unreachable") {
		t.Errorf("detail: %q", rows[0].Detail)
	}
}

func TestCheckGithubSealDisabledNoRepoCheck(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user": route(200, map[string]any{"login": "testuser"}),
	})
	rows := CheckGitHub(githubCfg(map[string]any{
		"vaultseal": map[string]any{"enabled": false, "repo_name": "backup"},
	}), srv.root(), 15.0)
	var names []string
	for _, r := range rows {
		names = append(names, r.Name)
	}
	if len(names) != 2 || names[0] != "GitHub token" || names[1] != "Vault seal" {
		t.Fatalf("names = %v", names)
	}
	if rows[1].Level != LevelInfo {
		t.Errorf("seal row level = %s, want info", rows[1].Level)
	}
	// no repo call happened
	captured := srv.captured()
	if len(captured) != 1 || captured[0].Path != "/user" {
		t.Errorf("requests = %+v, want just /user", captured)
	}
}

func TestCheckGithubSealEnabledNoRepoNameIsInfoHint(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user": route(200, map[string]any{"login": "testuser"}),
	})
	rows := CheckGitHub(githubCfg(map[string]any{
		"vaultseal": map[string]any{"enabled": true, "repo_name": ""},
	}), srv.root(), 15.0)
	if len(rows) != 2 || rows[1].Name != "Vault repo" {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[1].Level != LevelInfo {
		t.Errorf("level = %s, want info", rows[1].Level)
	}
	if !strings.Contains(rows[1].Detail, "remembered like credentials") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
}

func TestCheckGithubWebsitesRepoCheckedWhenPipelineOn(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user":                        route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup":       route(200, map[string]any{"private": true}),
		"/repos/testuser/websites-dir": route(200, map[string]any{"private": true}),
	})
	rows := CheckGitHub(githubCfg(map[string]any{
		"pipelines":         map[string]any{"websites": true},
		"website_repo_name": "websites-dir",
	}), srv.root(), 15.0)
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	if rows[2].Name != "Websites repo" || rows[2].Level != LevelOK {
		t.Errorf("row 2 = (%s, %s)", rows[2].Name, rows[2].Level)
	}
}

func TestCheckGithubPersonalRepoNoPipelineGate(t *testing.T) {
	// v0.84.0 — the Personal vault's backup repo joins the check with
	// no pipeline gate, exactly like its seal
	srv := newJSONServer(t, map[string][2]any{
		"/user":                    route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup":   route(200, map[string]any{"private": true}),
		"/repos/testuser/personal": route(200, map[string]any{"private": true}),
	})
	rows := CheckGitHub(githubCfg(map[string]any{
		"manual_repo_name": "personal",
	}), srv.root(), 15.0)
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	if rows[2].Name != "Personal repo" || rows[2].Level != LevelOK {
		t.Errorf("row 2 = (%s, %s)", rows[2].Name, rows[2].Level)
	}
}

func TestCheckGithubRepoCheckedWarnCarriesStatus(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user":                  route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup": route(500, map[string]any{"message": "boom"}),
	})
	rows := CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	if rows[1].Level != LevelWarn {
		t.Fatalf("level = %s, want warn", rows[1].Level)
	}
	if !strings.Contains(rows[1].Detail, "could not be checked (HTTP 500)") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
}

// THE UA TABLE (2/3): _github_get speaks
// "GitCurator-connection-check" with the Bearer token and the GitHub
// Accept header — pinned from the wire.
func TestCheckGithubUATableEntry(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/user":                  route(200, map[string]any{"login": "testuser"}),
		"/repos/testuser/backup": route(200, map[string]any{"private": true}),
	})
	_ = CheckGitHub(githubCfg(nil), srv.root(), 15.0)
	captured := srv.captured()
	if len(captured) != 2 {
		t.Fatalf("requests = %d, want 2", len(captured))
	}
	for i, req := range captured {
		if req.UA != "GitCurator-connection-check" {
			t.Errorf("request %d UA = %q", i, req.UA)
		}
		if req.Auth != "Bearer tok" {
			t.Errorf("request %d Auth = %q, want 'Bearer tok'", i, req.Auth)
		}
		if req.Accept != "application/vnd.github+json" {
			t.Errorf("request %d Accept = %q", i, req.Accept)
		}
		if req.Method != "GET" {
			t.Errorf("request %d method = %s", i, req.Method)
		}
	}
	if captured[1].Path != "/repos/testuser/backup" {
		t.Errorf("repo path = %q", captured[1].Path)
	}
}
