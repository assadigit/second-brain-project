package vaultscan

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/prompts"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ReadPrompt mirrors _read_prompt: the scan prompt — s01_vaultscan.txt
// for the Websites vault, s02_reposscan.txt for the GitHub Repos
// vault (read directly; the reference returns ” when the file cannot
// be read — the Go door's error reads the same way).
func ReadPrompt(kind string) string {
	name := "s01_vaultscan.txt"
	if isReposKind(kind) {
		name = "s02_reposscan.txt"
	}
	text, err := prompts.LoadPromptFile(name)
	if err != nil {
		return ""
	}
	return text
}

// LLMProposal mirrors _llm_propose's merged answer:
// {'new_folders', 'moves', 'summary', 'reviewed'}.
type LLMProposal struct {
	NewFolders []string
	Moves      []RawMove
	Summary    string
	Reviewed   []string
}

// LLMCallFn mirrors the injected llm_call: messages + task → the raw
// answer (a panicking call is survived exactly like the reference's
// raising one — the pass stops asking, the landed chunks survive).
type LLMCallFn = func(messages []llmclient.Message, task string) string

// llmPropose mirrors _llm_propose (the LLM pass: folder tree +
// candidate digests → the filing proposal JSON): the chunked filing
// pass (v0.70.0) — candidates ride the LLM in batch-sized chunks, the
// chunks share one tree, a broken or unparseable answer drops only
// that chunk, and un-proposed notes in landed chunks ride 'reviewed'
// (the census law).
func llmPropose(inv *Inventory, candidates []*NoteRow, llmCall LLMCallFn,
	log LogFn, tooBroad, maxNotes, batch int, progress ProgressFn,
	kind string) LLMProposal {
	prompt := ReadPrompt(kind)
	if prompt == "" || len(candidates) == 0 {
		merged := LLMProposal{
			NewFolders: []string{},
			Moves:      []RawMove{},
			Reviewed:   []string{},
		}
		if len(candidates) == 0 {
			merged.Summary = "nothing to file"
		}
		return merged
	}
	treeLines := FolderTreeLines(inv, tooBroad)
	tree := strings.Join(treeLines, "\n")
	if tree == "" {
		tree = "(empty)"
	}
	shown := candidates
	if len(shown) > maxNotes {
		shown = shown[:maxNotes]
	}
	chunk := batch
	if chunk < 1 {
		chunk = DefaultLLMBatch
	}
	var chunks [][]*NoteRow
	for i := 0; i < len(shown); i += chunk {
		end := i + chunk
		if end > len(shown) {
			end = len(shown)
		}
		chunks = append(chunks, shown[i:end])
	}
	totalChunks := len(chunks)
	repos := isReposKind(kind)
	prog := func(done int) {
		if progress == nil {
			return
		}
		func() {
			defer func() { recover() }() // the reference's try/except
			progress("LLM filing pass", done, totalChunks)
		}()
	}
	merged := LLMProposal{NewFolders: []string{}, Moves: []RawMove{}}
	landedChunks := 0
	for idx, part := range chunks {
		partRows := make([]NoteRow, len(part))
		for i, pr := range part {
			partRows[i] = *pr
		}
		digest := strings.Join(DigestLines(partRows, kind), "\n")
		if idx == totalChunks-1 && len(candidates) > len(shown) {
			digest += fmt.Sprintf("\n… +%d more candidate notes "+
				"(summarized away — plan for the listed ones only)",
				len(candidates)-len(shown))
		}
		passWord := ""
		if totalChunks > 1 {
			passWord = fmt.Sprintf(", pass %d of %d", idx+1, totalChunks)
		}
		coreWord := "the Website Directory"
		if repos {
			coreWord = "the vault's own directory notes"
		}
		confidenceWord := "domain, tags or declared home"
		homesPara := "settled siblings of its domain, " +
			"tags shared with the notes already filed there, title " +
			"words in common, its declared category/subcategory, the " +
			"taxonomy's vocabulary, a folder-name match)"
		if repos {
			confidenceWord = "owner, language, tags or declared category"
			homesPara = "tags shared with the notes already filed there, " +
				"title words in common, its declared category, a " +
				"folder-name match)"
		}
		userMsg := fmt.Sprintf(
			"Folder tree (relative to the vault root) — THE LATEST STATE, "+
				"checked fresh this run; read it before proposing any folder:\n"+
				"%s\n\n"+
				"Move-candidate notes (%d of %d%s):\n"+
				"%s\n\n"+
				"The three laws, again, at the point of decision: (1) check "+
				"the tree above for the same or a near-identical existing "+
				"folder before proposing ANY new one — reuse it instead; "+
				"(2) every destination is at least Category/Subfolder deep — "+
				"never a bare category root, never the vault root; (3) core "+
				"catalog notes (%s) are never moved.\n\n"+
				"Each move carries your own confidence: \"high\" when the "+
				"note's title, %s clearly name "+
				"its topic and the destination is obvious; \"low\" when you "+
				"are guessing (thin signals, ambiguous topic, a tag that "+
				"could name two homes). An honest low-confidence move is "+
				"SHOWN to the owner un-ticked — he decides; a wrong "+
				"high-confidence move is applied by default. When in doubt, "+
				"say \"low\" — or leave the note out entirely.\n\n"+
				"Some candidate lines end with a \"homes:\" shortlist — "+
				"up to three EXISTING folders the vault's own evidence "+
				"weighed for this note (%s, each with a "+
				"0.0-1.0 weight and the reasons that earned it. The "+
				"weights are ADVICE, not law: file into a shortlisted "+
				"home when it fits; override the arithmetic when you see "+
				"something the numbers cannot (say why in the move's "+
				"reason); propose a new folder only when nothing fits. A "+
				"CLOSE CALL marker means the top two homes weigh nearly "+
				"the same — you, not the arithmetic, break the tie. When you "+
				"override a home the evidence weighed clearly higher, "+
				"be honest and mark that move's confidence \"low\" — it "+
				"rides the plan un-ticked for the owner either way.\n\n"+
				"Return the JSON filing plan now.",
			tree,
			len(part), len(candidates), passWord,
			digest,
			coreWord,
			confidenceWord,
			homesPara)
		if totalChunks > 1 {
			log(fmt.Sprintf("🧠 Filing pass %d of %d — %d candidate "+
				"note(s) to the LLM…", idx+1, totalChunks, len(part)), "info")
		}
		prog(idx)
		var raw string
		crashed := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					crashed = true
				}
			}()
			raw = llmCall([]llmclient.Message{
				{Role: "system", Content: prompt},
				{Role: "user", Content: userMsg},
			}, "vaultscan")
		}()
		if crashed {
			log(fmt.Sprintf("⚠️ The scan's LLM call failed (pass %d of "+
				"%d): the guest panicked — the filing plan for those "+
				"notes is dropped this run (the deletions below are "+
				"unaffected)", idx+1, totalChunks), "warning")
			break
		}
		data, err := extractJSONDict(raw)
		if err != nil {
			log(fmt.Sprintf("⚠️ The scan's LLM answer was not JSON I "+
				"could read (pass %d of %d: %v) — that chunk's filing "+
				"plan is dropped this run", idx+1, totalChunks, err),
				"warning")
			continue
		}
		landedChunks++
		// v0.78.0 — the census's honest review list: every note this
		// LANDED chunk showed the guest that the guest did NOT propose
		// a move for.
		proposed := map[string]bool{}
		for _, m := range data.Moves {
			note := m.Note
			if note == "" {
				note = m.File
			}
			proposed[helpers.Strip(note)] = true
		}
		for _, n := range part {
			if !proposed[n.File] {
				merged.Reviewed = append(merged.Reviewed, n.Path)
			}
		}
		merged.NewFolders = append(merged.NewFolders, data.NewFolders...)
		for _, m := range data.Moves {
			note := m.Note
			if note == "" {
				note = m.File
			}
			merged.Moves = append(merged.Moves, RawMove{
				Note: note, To: m.To, Reason: m.Reason,
				Confidence: m.Confidence,
			})
		}
		if pyTruthyAny(data.Summary) {
			merged.Summary = cutRunesStr(strOf(data.Summary), 400)
		}
	}
	prog(totalChunks)
	if totalChunks > 1 && landedChunks < totalChunks {
		log(fmt.Sprintf("🧠 The filing pass landed %d of %d LLM "+
			"answer(s) — the missing chunks' notes stay put and ride "+
			"the next scan", landedChunks, totalChunks), "warning")
	}
	return merged
}

// rawLLMMove mirrors the guest's raw move dict (note/file, to, reason,
// confidence — plus the stray keys the validator ignores).
type rawLLMMove struct {
	Note       string
	File       string
	To         string
	Reason     string
	Confidence string
}

// rawLLMData mirrors the guest's raw proposal dict.
type rawLLMData struct {
	NewFolders []string
	Moves      []rawLLMMove
	Summary    any
}

// extractJSONDict runs llmclient.ExtractJSON and shapes the answer as
// the proposal dict (the house JSON shapes — storage.OrderedMap and
// scalar maps — handled per the T-42 law). A payload that is not a
// JSON OBJECT is a parse failure.
func extractJSONDict(content string) (rawLLMData, error) {
	var out rawLLMData
	v, err := llmclient.ExtractJSON(content)
	if err != nil {
		return out, err
	}
	obj := map[string]any{}
	switch t := v.(type) {
	case map[string]any:
		obj = t
	case *storage.OrderedMap:
		for _, k := range t.Keys() {
			if val, ok := t.Get(k); ok {
				obj[k] = val
			}
		}
	default:
		return out, fmt.Errorf("response is not a JSON object")
	}
	if nf, ok := obj["new_folders"]; ok {
		if list := asList(nf); list != nil {
			for _, item := range list {
				out.NewFolders = append(out.NewFolders, helpers.PyStr(item))
			}
		}
	}
	if mv, ok := obj["moves"]; ok {
		if list := asList(mv); list != nil {
			for _, item := range list {
				m := asDict(item)
				if m == nil {
					continue
				}
				out.Moves = append(out.Moves, rawLLMMove{
					Note:       strOf(m["note"]),
					File:       strOf(m["file"]),
					To:         strOf(m["to"]),
					Reason:     strOf(m["reason"]),
					Confidence: strOf(m["confidence"]),
				})
			}
		}
	}
	if s, ok := obj["summary"]; ok {
		out.Summary = s
	}
	return out, nil
}

// pyTruthyAny mirrors Python truthiness for the JSON-decoded shapes
// (nil, false, 0, ”, empty list/dict are falsy).
func pyTruthyAny(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case int64:
		return t != 0
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	case *storage.OrderedMap:
		return t.Len() > 0
	}
	return true
}

func asList(v any) []any {
	switch t := v.(type) {
	case []any:
		return t
	}
	return nil
}

func asDict(v any) map[string]any {
	switch t := v.(type) {
	case map[string]any:
		return t
	case *storage.OrderedMap:
		m := make(map[string]any, t.Len())
		for _, k := range t.Keys() {
			if val, ok := t.Get(k); ok {
				m[k] = val
			}
		}
		return m
	}
	return nil
}

func strOf(v any) string {
	if v == nil {
		return ""
	}
	return helpers.PyStr(v)
}

func cutRunesStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
