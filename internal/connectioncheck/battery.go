package connectioncheck

// battery.go — run_local_checks (RE-SCOPED per ADR-0004 Am.3/R-0:
// Vaults, LLM, GitHub, and the Worker leg promoted to
// queue-reachability — the Telegram section is dropped) + summarize
// (the re-worded headline: the Python headline's "Telegram" mention
// goes; the sync queue rides the Worker).

import (
	"fmt"
)

// Section is one battery group: [title, results] (the results slice
// stays appendable — orchestrators add live legs to the last section).
type Section struct {
	Title   string
	Results []Result
}

// panicAsError renders a recovered panic as an error (the reference's
// "check crashed ({Type}: {e})" needs both a class token and a
// message; Go panics carry values, not exceptions).
func panicAsError(r any) error {
	return fmt.Errorf("%v", r)
}

// panicClass renders the crash-row class token: Go's %T base name for
// error/any values (Python renders the exception class name; the
// shape — "check crashed ({class}: {msg})" — is the pinned contract).
func panicClass(r any) string {
	s := fmt.Sprintf("%T", r)
	if i := len(s) - 1; i >= 0 {
		for j := i; j >= 0; j-- {
			if s[j] == '.' {
				return s[j+1:]
			}
		}
	}
	return s
}

// RunLocalChecks runs the four sections in order: Vaults, LLM, GitHub,
// Worker. Each check is guarded — a crash degrades to a
// "check crashed" error row, never a panic. onSection receives
// (title, index, total); onResult receives each result; panics inside
// the callbacks are swallowed (the reference swallows callback
// exceptions). Never raises.
func RunLocalChecks(config any,
	onSection func(title string, idx, total int),
	onResult func(Result)) []Section {

	groups := []struct {
		title string
		fn    func() []Result
	}{
		{"Vaults", func() []Result { return CheckVaults(config) }},
		{"LLM", func() []Result { return CheckLLM(config) }},
		{"GitHub", func() []Result {
			return CheckGitHub(config, GitHubAPIBase, 15.0)
		}},
		{"Worker", func() []Result {
			wc := CheckWorkerVersion(config, nil)
			if wc == nil {
				return nil
			}
			return []Result{*wc}
		}},
	}
	sections := make([]Section, 0, len(groups))
	total := len(groups)
	for idx, g := range groups {
		title := g.title
		results := guardedResults(title, g.fn)
		sections = append(sections, Section{Title: title, Results: results})
		if onSection != nil {
			func() {
				defer swallowPanic()
				onSection(g.title, idx+1, total)
			}()
		}
		if onResult != nil {
			for _, r := range results {
				rr := r
				func() {
					defer swallowPanic()
					onResult(rr)
				}()
			}
		}
	}
	return sections
}

// guardedResults runs one section's check with the crash law: a
// panicked check yields the section's single error row (named for the
// section, exactly like the reference's "check crashed" row).
func guardedResults(title string, fn func() []Result) (results []Result) {
	defer func() {
		if r := recover(); r != nil {
			results = []Result{result(title, LevelError,
				"check crashed ("+panicClass(r)+": "+panicAsError(r).Error()+")")}
		}
	}()
	return fn()
}

func swallowPanic() {
	_ = recover()
}

// Summary is summarize()'s verdict: one level per section, then one
// headline.
type Summary struct {
	Level    string `json:"level"`
	Headline string `json:"headline"`
	Ready    int    `json:"ready"`
	Total    int    `json:"total"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
}

// Summarize reduces the sections to one verdict. All-ok headline:
// "ALL SYSTEMS READY — vaults, LLM, GitHub and the sync queue are up"
// (the re-worded headline — the Python's "Telegram" mention goes per
// ADR-0004 Am.3; the Worker leg IS the queue-reachability). Anything
// else: "{ready}/{total} subsystems ready — N error(s) and M
// warning(s) (see the lines above)". Empty/nil sections are safe
// (level ok — the reference's empty-section law).
func Summarize(sections []Section) Summary {
	per := []string{}
	for _, sec := range sections {
		levels := []string{}
		for _, r := range sec.Results {
			levels = append(levels, r.Level)
		}
		if len(levels) == 0 {
			levels = []string{LevelInfo}
		}
		switch {
		case containsStr(levels, LevelError):
			per = append(per, LevelError)
		case containsStr(levels, LevelWarn):
			per = append(per, LevelWarn)
		default:
			per = append(per, LevelOK)
		}
	}
	ready, errs, warns := 0, 0, 0
	for _, lv := range per {
		switch lv {
		case LevelOK:
			ready++
		case LevelError:
			errs++
		case LevelWarn:
			warns++
		}
	}
	level := LevelOK
	if errs > 0 {
		level = LevelError
	} else if warns > 0 {
		level = LevelWarn
	}
	s := Summary{
		Level: level, Ready: ready, Total: len(per),
		Errors: errs, Warnings: warns,
	}
	if level == LevelOK {
		s.Headline = "ALL SYSTEMS READY — vaults, LLM, GitHub and " +
			"the sync queue are up"
		return s
	}
	bits := []string{}
	if errs > 0 {
		bits = append(bits, itoa(errs)+" error(s)")
	}
	if warns > 0 {
		bits = append(bits, itoa(warns)+" warning(s)")
	}
	s.Headline = itoa(ready) + "/" + itoa(len(per)) +
		" subsystems ready — " + joinAnd(bits) +
		" (see the lines above)"
	return s
}

// joinAnd is Python's " and ".join.
func joinAnd(bits []string) string {
	out := ""
	for i, b := range bits {
		if i > 0 {
			out += " and "
		}
		out += b
	}
	return out
}
