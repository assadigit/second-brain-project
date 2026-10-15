# webui/ — the Wails GUI shell (Phase 6)

The Wails frontend that replicates the documented PyQt6 design lives here
from Phase 6 onward (ADR-0001: workflow-preserving HTML replica; no visual
redesign — `docs/port-blueprint/07-ui-and-performance-assessment.md` Part D
marks the boundary).

During Phase 0, the Wails proof-of-concept (task T-70: themed window +
log desk + one worker job) lives in `poc/wails/` as a standalone module so
the main module stays dependency-free until the ADR verdicts are recorded.
