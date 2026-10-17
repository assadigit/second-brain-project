package websitepipeline

import (
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// The module constants (website_pipeline.py:76-87 + the verdict
// grammar's marker sets — the table-marker constants themselves live
// in decomtable, the shared grammar floor).

// WebsitePromptVersion mirrors WEBSITE_PROMPT_VERSION: bump when the
// website prompt set changes shape (SPEC Appendix A). web-v2
// (v0.27.0): decision-oriented body — "What it does" replaces the
// marketing-prone "Core offerings"/"Standout feature" pair, unknown
// values are omitted instead of printed as "unknown", and two new
// sections carry the practical facts and honest caveats. "Best used
// for" keeps its exact heading (recall/linking parses it) and
// "Similar tools" stays the closing section.
const WebsitePromptVersion = "web-v2"

// ReviewFolder mirrors REVIEW_FOLDER.
const ReviewFolder = decomtable.ReviewFolder

// DecommissionTable mirrors DECOMMISSION_TABLE (re-exported view of
// the decomtable constant — the pipeline's own name for it).
const DecommissionTable = decomtable.DecommissionTable

// FetchTimeoutS mirrors FETCH_TIMEOUT_S.
const FetchTimeoutS = 20.0

// FetchMaxBytes mirrors FETCH_MAX_BYTES.
const FetchMaxBytes = 2_000_000

// DomainDelayS mirrors DOMAIN_DELAY_S.
const DomainDelayS = 2.0

// ClassifyRetries mirrors CLASSIFY_RETRIES: SPEC §4.6 — retry up to 2
// times, then _review.
const ClassifyRetries = 2

// LLMExcerptChars mirrors LLM_EXCERPT_CHARS: page text given to the
// model.
const LLMExcerptChars = 4000

// MinTextForAnalysis mirrors MIN_TEXT_FOR_ANALYSIS: below this, the
// model gets title+desc only.
const MinTextForAnalysis = 80

// LowConfidence mirrors LOW_CONFIDENCE: confidence that routes to
// _review.
const LowConfidence = "low"

// HandDoorHint mirrors HAND_DOOR_HINT (v0.48.0): the fourth door's
// suffix — a failure whose class the three machine doors could not
// open gains this line, so the _review note, the master table's Notes,
// and the log all name the door that CAN answer it.
const HandDoorHint = " — the fourth door: 🖐 hand-deliver it (More ▸ Hand-deliver walled links opens the link in your real Chrome; save the page into the hand-delivered folder and the next batch takes it from there)"

// SiteLinksHeading mirrors SITE_LINKS_HEADING (v0.64.0 — ONE NOTE PER
// SITE): the body section every consolidated note carries (the human's
// half of the law — every extra link of the site, one list, one note).
const SiteLinksHeading = "## Links on this site"

// TrashFolderNames mirrors TRASH_FOLDER_NAMES (v0.62.0 — THE TRASH
// DOOR): a root-level folder of this name (any spelling) reads as the
// owner's delete verdict for notes moved into it.
var TrashFolderNames = []string{"trash"}

// TrashMarker mirrors TRASH_MARKER: the log/record label the trash
// door's items carry.
const TrashMarker = "Trash folder"

// TrashGesture mirrors TRASH_GESTURE: the gesture label.
const TrashGesture = "Trash folder move"

// BanishReasonPrefix mirrors BANISH_REASON_PREFIX: the dismissal
// reason prefix the skip gate reads (process_link names the door so
// the log tells a banishment from a graveyard burial).
const BanishReasonPrefix = "banished by owner"

// BanishTallyPrefix mirrors BANISH_TALLY_PREFIX (v0.59.0 — THE TALLY):
// the log line every run answers with.
const BanishTallyPrefix = "🗑️ Run tally:"

// BanishQuarantineRelpath mirrors BANISH_QUARANTINE_RELPATH: where
// banished notes go — inside Obsidian's hidden .trash (out of the
// library, invisible to VaultIndex/mirror/directory, recoverable by
// hand; ♻️ revived + a hand move brings a mistaken burial back).
const BanishQuarantineRelpath = ".trash/banished"

// BanishGatePrefix mirrors BANISH_GATE_PREFIX (v0.60.0 — THE
// CONFIRMATION GATE): the log prefix of the run's opening ask.
const BanishGatePrefix = "🗑️ Vault scan:"

// BanishGateEndpoint mirrors BANISH_GATE_ENDPOINT: the state-table row
// the worker keeps the ask in (banish_confirm:<id> —
// pending/confirmed/declined/timeout).
const BanishGateEndpoint = "/api/banish"

// BodyTagStops mirrors _BODY_TAG_STOPS (v0.60.1): the body-tag scan's
// stop set — an Obsidian inline tag ends at whitespace, another #, or
// any of these (#delete. is the tag delete plus sentence punctuation).
const BodyTagStops = " \t\r\n,.;:!?)]}'\"»«…—>"

// ProxyEpochKey mirrors WebsitePipeline.PROXY_EPOCH_KEY: the meta key
// the proxy-epoch retry re-arm rides.
const ProxyEpochKey = "web_proxy_epoch"
