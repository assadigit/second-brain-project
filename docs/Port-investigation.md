# Mission: Reverse-Engineer the Python Application and Produce a Complete Go Migration Blueprint

## 1. Your Role

Act as a senior software architect, Python engineer, Go engineer, desktop application engineer, performance analyst, and technical documentation specialist.

You have access to an existing software repository. The application is currently written in Python, and the codebase is relatively small, approximately fewer than 20,000 lines including comments.

Your task is to inspect the actual repository, understand how the application works, and create a complete, implementation-ready blueprint for migrating it from Python to the Go programming language.

**This is a planning, reverse-engineering, and specification-writing session. It is NOT an implementation session.**

A separate AI-agent session will use the documentation you produce to implement the migration. That future agent must be able to understand what needs to be built, why it needs to be built, how its behavior must match the existing application, which technical decisions have been made, and how to verify the result.

Do not assume knowledge of our previous conversation. Everything the next agent needs must be documented in the repository.

## 2. Project Goals and Constraints

Use the following requirements as the basis for all your analysis and recommendations.

### Primary goal: A reliable Go migration

The future Go application must preserve the Python application's existing functionality and externally observable behavior, unless a clearly documented, intentional improvement is approved by the project specifications.

The migration should aim to preserve:

- Existing features and supported workflows.
- Inputs, outputs, commands, and public interfaces.
- Data formats, configuration, settings, and persistence behavior.
- Business rules, calculations, and processing logic.
- Error handling, validation, and important edge cases.
- Communication with external services and dependencies.
- Existing file operations and integration behavior.
- User-facing functionality and expected workflow.

Do not assume that the Python implementation is perfectly documented or that its intended behavior always matches its actual behavior.

The source code, tests, configuration, documentation, and observable execution behavior are the evidence from which you must derive the specification.

### Secondary goal: Better performance

The current application feels too slow, particularly in its interface. The goal is to make the resulting application feel responsive and perform well.

Do not automatically conclude that rewriting the application in Go will solve every performance problem.

Investigate the current architecture and identify plausible causes of sluggishness, such as blocking operations, excessive rendering, inefficient polling, unnecessary repeated computations, synchronous network calls, slow startup, or inefficient data processing.

Distinguish confirmed bottlenecks from hypotheses. If reliable measurements cannot be taken in this session, specify how the implementation session should establish a baseline and test improvements.

Performance must improve without silently sacrificing correctness or functionality.

### Platform goal: Windows first

The first release should prioritize a reliable Windows executable that is straightforward for an ordinary user to launch.

The eventual deliverable should ideally be a distributable `.exe` with minimal setup, unnecessary dependencies, or manual configuration.

However, do not promise that a single executable can eliminate every runtime requirement. Some GUI technologies, native libraries, assets, and integrations may introduce additional packaging requirements.

Prefer architectural choices that keep future Linux and macOS support feasible where practical, without compromising the primary Windows goal.

### GUI goal: Preserve functionality now, redesign later

The user-facing workflow and functionality must remain consistent throughout the initial migration.

A different GUI technology may be appropriate, but a complete visual redesign is outside the scope of this session. Do not create new screen designs or implement a new interface.

Instead, assess the existing GUI technology, its architecture, performance limitations, and realistic migration options. Recommend a suitable long-term GUI strategy for the implementation phase.

A future UI redesign can be planned separately once the Go application has a reliable functional foundation.

### Scope of this session

You must produce the complete migration blueprint, including architecture specifications, behavioral contracts, dependency replacement decisions, per-module specifications, testing strategy, build and distribution requirements, implementation tasks, risks, and acceptance criteria.

**Do not port the code, implement replacement modules, or redesign the actual application in this session.**

Your only intended repository changes are the planning, specifications, and documentation inside the `port-blueprint/` directory.

---

## 3. Repository Safety Rules

Before beginning the analysis:

1. Confirm that you are at the repository root.
2. Inspect the repository structure, current Git branch, and existing Git status.
3. Identify existing documentation, source code, tests, configuration files, assets, dependency declarations, and build scripts.
4. Check whether `port-blueprint/` already exists. If it does, inspect its existing contents and preserve relevant work instead of blindly overwriting it.

Strict rules:

- Do not modify application source code.
- Do not modify existing tests, application configuration, dependency manifests, lockfiles, build pipelines, or release files.
- Do not rename, delete, move, or reformat existing project files.
- Do not create experimental implementation code outside `port-blueprint/`.
- Do not install or upgrade dependencies merely to conduct the analysis.
- You may run safe, existing tests and non-destructive inspection commands when useful.
- Avoid commands that could overwrite user data, alter the Git history, or discard existing changes.
- Do not revert existing Git changes, even if they appear unrelated.
- Avoid introducing generated caches, temporary files, or unrelated modifications. Check the final Git status.
- Never include secrets, access tokens, passwords, private keys, or other credentials in the documentation.

If some analysis requires unavailable dependencies, credentials, services, operating systems, or tools, record the limitation and explain what remains unverified.

Do not modify anything outside `port-blueprint/`. If the blueprint folder already contains useful material, make careful, targeted updates inside that folder only.

---

## 4. Required Investigation Process

Conduct the analysis systematically. Do not start writing generic architecture documents before understanding the repository.

### Phase A — Repository reconnaissance

Inspect all relevant repository contents, including:

- Python source files.
- Package and module structure.
- Classes, functions, entry points, and initialization code.
- Test suites and test fixtures.
- Dependency manifests and lockfiles.
- Configuration files and environment-variable usage.
- Documentation, README files, and examples.
- GUI resources, images, fonts, stylesheets, icons, and other assets.
- Build, packaging, and release scripts.
- GitHub Actions or other CI/CD workflows.
- Database schemas, migrations, local storage, caches, and user data files.
- API definitions, network integrations, subprocesses, and operating-system integrations.

Inspect source code, not only filenames, comments, or existing documentation.

Determine how the program actually starts, which parts execute first, how control flows through the application, and how its main workflows operate.

Identify dead code, duplicated behavior, undocumented assumptions, questionable abstractions, and potentially obsolete files. Label them as findings rather than deleting or changing anything.

Produce a complete inventory of relevant modules, components, and their relationships.

Because this is a relatively small codebase, aim to account for every production module. Do not skip a module simply because it appears unimportant or difficult to understand.

### Phase B — Establish the current application architecture

Reconstruct the existing architecture from the implementation.

Identify:

- Application entry points and lifecycle.
- Major components and their responsibilities.
- Dependencies and import relationships.
- Data flow between components.
- Core business logic.
- GUI-to-backend communication.
- State management and persistence.
- Background work, threading, asynchronous execution, and event handling.
- Network requests, filesystem access, and external integrations.
- Error propagation and logging.
- Configuration and platform-dependent behavior.

Create architecture diagrams when useful. Mermaid diagrams are preferred for relationships and flows that can be represented clearly.

Do not force the existing implementation into an assumed architectural pattern. Describe what exists before recommending what should change.

### Phase C — Reverse-engineer behavior

This is one of the most important parts of the mission.

Document what the application does, not merely what its functions are called.

For each significant feature and workflow, determine:

- How the workflow starts.
- Which inputs it accepts.
- What validations are performed.
- How inputs are transformed and processed.
- Which internal components participate.
- Which outputs are generated or displayed.
- Which state changes occur.
- Which files, databases, or external services are affected.
- What happens when the operation succeeds.
- What happens when it fails.
- What happens with empty, invalid, malformed, missing, or boundary-value inputs.
- Whether execution is synchronous, asynchronous, or concurrent.
- What users can observe at each stage.

Pay special attention to behavior that can be accidentally changed during a language migration:

- Number precision, rounding, numeric conversions, and integer overflow.
- String handling, Unicode, character encodings, and whitespace.
- Date/time parsing, time zones, locale-dependent formatting, and serialization.
- Sorting, ordering, iteration, and deterministic output.
- Missing values, `None`, zero values, optional fields, and defaults.
- Python truthiness and implicit conversions.
- Exception types, error messages, and error propagation.
- File paths, relative versus absolute paths, permissions, and filesystem differences.
- JSON, YAML, CSV, database, and other serialization formats, where applicable.
- Floating-point calculations and other potentially sensitive numerical operations.
- Retries, timeouts, cancellation, concurrency, and resource cleanup.
- User-visible labels, status messages, notifications, and workflow transitions.
- Existing quirks that users may rely on, even when those quirks appear unintended.

Inspect relevant tests and fixtures to establish expected behavior. Where possible, distinguish behavior verified by tests from behavior inferred from implementation.

Do not invent intended behavior where the source is ambiguous. Record the ambiguity and create an explicit decision item.

### Phase D — Identify the real performance constraints

Investigate the current application's responsiveness and execution behavior.

Identify the likely causes of perceived slowness, considering both GUI and non-GUI work.

For each suspected bottleneck, record:

1. The relevant module, function, event handler, or execution path.
2. Evidence supporting the concern.
3. The likely impact on responsiveness.
4. Whether the problem is confirmed or hypothetical.
5. A recommended measurement or profiling method.
6. A proposed improvement and its functional risks.

Where practical and safe, collect reproducible baseline measurements. Do not fabricate performance numbers or claim an improvement without evidence.

If the application cannot be executed in the current environment, document the commands and measurement procedure needed for the implementation session.

The eventual implementation should have performance acceptance criteria that are measurable where possible, rather than vague requirements such as "make it faster."

---

## 5. Python-to-Go Migration Analysis

Determine how each existing component should be represented in the Go implementation.

Do not translate Python syntax mechanically into Go syntax. Design the Go version to be maintainable and idiomatic while preserving required behavior.

### Target architecture

Propose a concrete, repository-specific Go architecture.

Specify:

- A suitable Go module name strategy, based on available repository information.
- The recommended Go version or version-selection policy.
- A package and directory structure.
- Responsibilities and boundaries for each package.
- How domain logic should be separated from GUI, persistence, networking, and operating-system integrations where such separation is beneficial.
- How components communicate.
- Which interfaces are needed and where they belong.
- How dependencies are injected or constructed.
- How errors are represented and propagated.
- How logging, configuration, cancellation, and lifecycle management work.
- How testing can be performed without unnecessarily launching the GUI or accessing real external services.
- How to avoid circular dependencies and excessive abstraction.

Make the architecture appropriate for this actual codebase. Do not introduce layers, interfaces, factories, or frameworks simply because they are fashionable.

Include a proposed directory tree and explain the responsibilities of the important files and packages.

### Python-to-Go mapping

Create an explicit mapping from the existing Python modules and responsibilities to proposed Go packages and files.

For every important component, explain:

- What the existing component does.
- Which behaviors it owns.
- Where those responsibilities should live in Go.
- Whether the component should be preserved, merged, split, replaced, or removed.
- Which interfaces or data types it needs.
- What behavior might change during the conversion.
- Which tests will establish parity.

Create a dedicated specification for each production module, or a well-justified module group when the code structure makes grouping clearer. Every production module must be accounted for in the inventory and mapping.

### Type and semantic compatibility

Identify Python features that require careful translation, including:

- Dynamic typing and runtime validation.
- Dictionaries, lists, tuples, sets, and custom objects.
- Dataclasses, enums, protocols, decorators, generators, or metaprogramming, where used.
- Async operations, threading, and background execution.
- Exception-based error handling.
- Dynamic imports or plugin discovery, where applicable.
- Python libraries whose behavior differs materially from potential Go replacements.

Specify the corresponding Go design, relevant type definitions, expected semantics, and compatibility risks.

Do not generate a mechanical translation of every function. Focus on preserving behavior and defining implementation requirements clearly enough that another agent can write the Go code correctly.

---

## 6. Dependency Replacement Strategy

Inventory every material third-party dependency and any important standard-library feature whose behavior affects the application.

For each dependency, document:

- Python package or library name.
- Where it is used.
- Which functionality is relied upon.
- Whether it is essential or incidental.
- The recommended Go equivalent, alternative design, or decision still required.
- Why the choice is appropriate.
- Any differences in semantics or feature coverage.
- Compatibility risks.
- Licensing considerations where relevant.
- Platform requirements.
- Build implications, including CGO or native libraries where applicable.
- Testing required to establish equivalent behavior.

Do not assume a library has a direct Go equivalent.

Where alternatives exist, compare realistic options and recommend a default. Prioritize correctness, maintainability, compatibility, and deployment simplicity over popularity alone.

If a dependency cannot be replaced safely without more evidence, explicitly mark the choice as unresolved and specify how to evaluate it.

If you consult external documentation to verify library support, release status, build requirements, or current APIs, prefer official documentation. Include source URLs and the date checked. If external research is unavailable, identify facts that the implementation session must verify rather than guessing.

---

## 7. GUI Technology and Performance Strategy

The application's GUI must be assessed, but **a UI redesign is not part of this mission**.

First, identify the existing GUI technology and how it communicates with the application logic. Determine which components are responsible for presentation, business logic, blocking work, event processing, and state updates.

Then assess feasible GUI strategies for a Go-based application.

Consider appropriate candidates from the current Go GUI ecosystem, based on the application's actual requirements. Potential candidates might include native or native-oriented frameworks, cross-platform desktop toolkits, and Go frameworks that use embedded web technologies. These are examples for evaluation, not predetermined choices.

Compare viable candidates using criteria such as:

- Windows user experience and visual quality.
- Packaging and runtime requirements.
- Startup time and interaction responsiveness.
- Cross-platform support.
- Compatibility with existing functionality.
- Ease of preserving the current workflow.
- Accessibility and keyboard navigation.
- Rendering and background-work model.
- Maturity, maintenance, licensing, and ecosystem.
- Support for required controls, tables, charts, forms, menus, notifications, and other features found in the repository.
- Resource usage and operational complexity.

For each serious candidate, explain the trade-offs and identify any requirements it cannot meet well.

Recommend a primary option, a reasonable fallback, and the key evidence behind the recommendation.

Do not select a GUI technology merely because it is written in Go. Do not assume that compiling to an executable automatically provides a native-looking interface or superior responsiveness.

Where the best choice depends on measurements or user experience that cannot be evaluated in this session, record the decision as provisional and design a small proof-of-concept task for the implementation phase.

**The outcome of this section must be a documented GUI migration strategy, not an implemented interface.**

---

## 8. Data, Configuration, and Compatibility

Document how the existing application stores, loads, exchanges, and transforms data.

Identify all applicable:

- Configuration formats and defaults.
- Environment variables.
- User preferences.
- Local data files.
- Databases and schema versions.
- Caches and temporary files.
- External API contracts.
- Import and export formats.
- Persistent state and saved sessions.
- Credentials or authentication mechanisms.
- Logs and diagnostic outputs.
- Inter-process communication formats.
- Network protocols and serialization rules.

For each format or contract, define the requirements the Go version must preserve.

If storage or configuration changes are proposed, document the migration mechanism and backward-compatibility strategy.

Do not assume that newly generated data can replace existing user data safely.

Define a plan for testing old fixtures and representative existing data against the Go implementation.

---

## 9. Testing and Behavioral Parity

Design a comprehensive testing strategy for the future Go application.

The purpose is not merely to prove that the new code runs. It is to show that the Go version behaves consistently with the current Python version while meeting the new performance and packaging goals.

Inspect the existing test suite, identify gaps, and propose the tests needed to close them.

Cover, where relevant:

- Unit tests.
- Characterization tests that capture actual existing behavior.
- Golden-file or fixture-based tests.
- Integration tests.
- Contract tests for external APIs and dependencies.
- Data serialization and compatibility tests.
- Error-path and boundary-condition tests.
- Concurrency and cancellation tests.
- GUI workflow and interaction tests.
- Windows-specific integration tests.
- Smoke tests for packaged executables.
- Regression tests for previously identified bugs.
- Performance benchmarks and profiling.
- Cross-platform tests if those platforms are supported.

### Python-versus-Go comparison

Where feasible, propose a differential testing approach: run equivalent test cases through the Python implementation and the Go implementation, then compare their observable results.

Define which results should match exactly and which may need a documented normalization, such as generated timestamps, nondeterministic identifiers, or platform-specific paths.

The blueprint must distinguish acceptable differences from genuine migration regressions.

Do not assume that the Go implementation should reproduce a bug merely because it exists. Classify suspected defects separately. Recommend preserving observable behavior by default, and require explicit documentation and regression tests for any intentional correction.

If an existing bug should be fixed, record the rationale, expected new behavior, and affected acceptance criteria.

Produce a traceability relationship between requirements, behavioral specifications, and tests.

---

## 10. Windows Build, Distribution, and Release Plan

Define how the future application should be built, packaged, tested, and distributed.

Address the following where applicable:

- Reproducible Go builds.
- Windows architecture selection, particularly the intended 64-bit target.
- Executable naming and output location.
- GUI subsystem configuration so a GUI application does not accidentally open an unwanted console window.
- Required DLLs, native libraries, web runtimes, assets, fonts, icons, and manifests.
- Whether CGO is necessary and how it affects build environments and portability.
- Whether the executable can genuinely run independently or requires documented runtime components.
- Clean-machine testing.
- File paths and permissions.
- Configuration and user-data locations.
- Application versioning.
- Logging and diagnostics.
- Installation versus portable distribution.
- Optional code signing and antivirus false-positive considerations.
- Automated build and release workflows.
- Future Linux and macOS build feasibility, if supported by the chosen architecture.

Do not assume that every GUI framework produces a fully self-contained executable. Document the selected framework's actual packaging requirements and outstanding verification tasks.

Propose reproducible build commands and a release-validation checklist based on the eventual architecture. Clearly label commands that still need to be validated during implementation.

---

## 11. Required Documentation and Deliverables

Create the following directory inside the repository:

`port-blueprint/`

Organize the deliverables as follows. Adapt filenames or split files where necessary to accurately reflect the repository, but preserve the complete coverage of these requirements.

### Main documents

`README.md`
- Explains the purpose and organization of the blueprint.
- Gives the next implementation agent an ordered reading guide.
- Identifies the recommended migration approach.
- Documents assumptions, known limitations, unresolved decisions, and the current status of the blueprint.
- Explains how to use the specifications without relying on this conversation.

`00-executive-summary.md`
- Concise description of the application and migration goals.
- Main architectural findings.
- Recommended target architecture and GUI approach.
- Major risks, trade-offs, and decisions requiring attention.

`01-repository-inventory.md`
- Repository structure.
- Production-module inventory.
- Entrypoints, tests, assets, configuration, dependencies, external integrations, and build infrastructure.
- Files or components whose roles remain uncertain.
- Coverage accounting for production code.

`02-current-architecture.md`
- Existing architecture and component relationships.
- Startup and shutdown behavior.
- Important execution and data flows.
- Diagrams where useful.
- Current architectural constraints and weaknesses.

`03-behavioral-specification.md`
- Observable application behavior.
- User workflows and system workflows.
- Functional requirements, edge cases, failure behavior, and compatibility requirements.
- References to detailed per-module specifications.

`04-target-go-architecture.md`
- Proposed Go package structure.
- Package responsibilities and dependency rules.
- Data types, interfaces, error handling, lifecycle, configuration, logging, concurrency, and application composition.
- Architecture diagrams and design rationale.

`05-python-to-go-mapping.md`
- Mapping of existing Python modules and responsibilities to their intended Go destinations.
- Major semantic conversion risks.
- A list of components that should be retained, replaced, consolidated, or redesigned internally.

`06-dependency-migration.md`
- Dependency inventory and replacement matrix.
- Recommended libraries and alternatives.
- Compatibility differences, licensing, platform constraints, and unresolved dependency decisions.

`07-ui-and-performance-assessment.md`
- Existing GUI architecture.
- Performance findings and hypotheses.
- GUI framework comparison and recommendation.
- Measurement strategy and provisional decisions.
- Explicit boundary between the migration phase and the later UI redesign.

`08-data-and-compatibility.md`
- Data formats, storage, configuration, persistence, API contracts, and backward compatibility.
- Required migration and preservation strategies.

`09-testing-and-parity.md`
- Test strategy, behavioral parity requirements, characterization tests, differential testing, regression coverage, and performance validation.

`10-build-release-packaging.md`
- Windows executable strategy.
- Packaging, runtime dependencies, clean-machine validation, build commands, release automation, and future platform considerations.

`11-security-reliability.md`
- Relevant security boundaries, sensitive-data handling, external input validation, resource cleanup, reliability risks, error reporting, and resilience requirements.
- Cover only the security and reliability issues applicable to the actual application.

`12-migration-roadmap.md`
- Ordered migration phases.
- Dependencies between phases.
- Milestones, entry conditions, completion conditions, and verification gates.
- Required prototypes or decisions that must precede implementation.

`13-implementation-task-list.md`
- A concrete and ordered set of implementation tasks.
- Each task must be specific, actionable, and traceable to one or more specifications.
- Identify prerequisites, affected Go packages or files, required tests, acceptance criteria, and relevant risks.
- Use relative effort categories such as XS, S, M, L, and XL where helpful. Do not invent precise implementation time estimates.

`14-risks-assumptions-open-questions.md`
- Confirmed risks, uncertainties, inferred assumptions, unresolved technical decisions, and recommended resolutions.
- Distinguish findings that require no further action from items that block or influence implementation.

`15-acceptance-criteria.md`
- Defines the overall completion conditions for the Go migration.
- Includes functionality, compatibility, correctness, responsiveness, testing, packaging, and documentation.
- Makes clear how the implementation team should demonstrate that the port is complete.

`16-next-session-implementation-prompt.md`
- A ready-to-copy prompt for the next AI coding-agent session.
- Instructs the next agent to read and follow the blueprint, verify its assumptions against the repository, implement the migration in the planned order, maintain tests, and demonstrate compatibility.
- Explicitly prohibits blindly trusting a specification when source-code evidence contradicts it.
- Requires the implementation session to work in controlled phases and report completed, pending, and blocked items.

### Detailed module specifications

Create:

`port-blueprint/module-specs/`

Add a detailed Markdown specification for every production Python module, or a clearly justified group of closely related modules.

Each specification should include, as applicable:

1. Source module and relevant symbols.
2. Responsibility and place in the architecture.
3. Current dependencies.
4. Inputs, outputs, state changes, and side effects.
5. Public interfaces and important internal contracts.
6. Business rules and transformations.
7. Validation, defaults, edge cases, and failure behavior.
8. Concurrency, lifecycle, and resource-management concerns.
9. Proposed Go package and file mapping.
10. Required types, interfaces, and dependency replacements.
11. Behavioral compatibility requirements.
12. Required unit, integration, and regression tests.
13. Implementation prerequisites and unresolved decisions.
14. Evidence references to the original source.

Specifications must be detailed enough that an implementation agent can implement a module without having to rediscover its behavior from scratch.

For a very small module, keep the documentation proportional to its complexity. Do not create empty boilerplate sections or repeat large amounts of text unnecessarily.

### Decision records

Create:

`port-blueprint/decisions/`

Use Architecture Decision Records (ADRs) for consequential design choices, such as the GUI framework, major dependency replacements, package boundaries, persistence strategy, or compatibility policy.

Each decision record should include:

- Context.
- The problem being solved.
- Options considered.
- The chosen option and rationale.
- Advantages and disadvantages.
- Consequences and risks.
- Evidence or external references.
- Status: proposed, accepted, provisional, or rejected.

Do not fabricate decisions that have not been sufficiently evaluated. Use provisional records when additional verification is required.

### Traceability matrix

Create:

`port-blueprint/traceability-matrix.md`

Connect the important requirements and behaviors to:

- Their source-code evidence.
- Relevant module specifications.
- Planned Go components.
- Test cases or required test categories.
- Implementation tasks.
- Acceptance criteria.

The purpose is to prevent important existing functionality from disappearing during migration.

---

## 12. Evidence, Accuracy, and Writing Standards

Every important finding must be grounded in repository evidence.

When referring to existing implementation details, cite the relevant repository-relative file path, symbol, and line range wherever practical.

For example:

`src/example.py:40-78 — ExampleProcessor.process()`

Verify line numbers against the actual source. Do not invent paths, functions, dependencies, tests, or current behavior.

Distinguish clearly between:

- **Observed:** directly supported by source code, tests, execution, or configuration.
- **Inferred:** a conclusion supported by available evidence but not directly confirmed.
- **Recommendation:** a proposed design or improvement.
- **Unresolved:** a question that cannot responsibly be settled with current evidence.

Use confidence labels when they help identify uncertain conclusions.

When a feature cannot be understood completely, document the exact missing evidence and its implications. Do not silently omit it or invent an answer.

Prefer concrete, repository-specific documentation over generic explanations of Python, Go, software architecture, or testing.

Write for two audiences: the technical AI agent that will implement the migration and a non-specialist project owner who needs to understand important decisions. Explain unfamiliar technical terms in plain language when doing so helps the owner make decisions.

Avoid excessive jargon, unexplained abbreviations, and unnecessary complexity.

---

## 13. Completeness and Quality Requirements

Before declaring the blueprint complete, verify all of the following.

### Repository coverage

- All production modules have been inventoried and mapped.
- All known entry points have been identified.
- All important workflows and observable behaviors are documented.
- Relevant tests, assets, configuration, dependencies, integrations, and build files have been inspected.
- Uncertain components are explicitly recorded.
- No significant behavior has been omitted merely because it is not well documented.

### Implementation readiness

- The target Go architecture is concrete and specific to this repository.
- Dependencies have recommendations or explicit unresolved decisions.
- Each module has a usable implementation specification.
- Compatibility-sensitive behavior has been documented.
- The testing strategy explains how parity will be established.
- The Windows executable and packaging strategy addresses real deployment requirements.
- Implementation tasks are ordered and have testable acceptance conditions.
- The next-session prompt is self-contained and usable.
- The documentation is internally consistent, and cross-references point to actual files.

### Performance and GUI readiness

- The cause of the current sluggishness has been investigated as far as the available evidence permits.
- Confirmed findings are distinguished from hypotheses.
- The blueprint describes how to measure performance.
- GUI technology has been evaluated against the application's actual needs.
- The recommendation accounts for Windows packaging and future portability.
- The migration phase is kept separate from the later visual redesign.

### Safety and repository integrity

- No application source code or existing project configuration was modified.
- No unrelated files were changed or deleted.
- Existing Git changes remain intact.
- The only intended persistent changes are the blueprint documents inside `port-blueprint/`.

Do not mark an item complete merely because a section exists. Confirm that it contains useful, evidence-based content.

---

## 14. Final Execution Procedure

Follow this sequence:

1. Inspect the repository and record the initial Git status.
2. Inventory the source code, tests, dependencies, architecture, data contracts, and GUI.
3. Reconstruct observable behavior and document compatibility requirements.
4. Investigate performance concerns and possible GUI migration strategies.
5. Design the target Go architecture and dependency replacements.
6. Write detailed module specifications and behavior contracts.
7. Design the testing, parity, packaging, and release strategies.
8. Create the ordered roadmap and implementation task list.
9. Generate the traceability matrix and next-session implementation prompt.
10. Cross-check the documentation against the source code.
11. Validate document links, file references, and internal consistency.
12. Inspect the final Git status to ensure that nothing outside `port-blueprint/` was modified.

If you encounter uncertainty, resolve it using repository evidence or record it as an open question with a recommended verification procedure. Do not stop the entire mission merely because one dependency or design choice remains uncertain.

Prioritize useful, honest, complete specifications over the appearance of certainty.

## 15. Final Report

Once the blueprint documents have been written and checked, provide a concise report containing:

- The high-level findings about the application.
- The recommended migration architecture.
- The recommended GUI direction and the basis for that recommendation.
- The most important performance observations.
- Major compatibility risks and unresolved decisions.
- The documentation files created.
- The tests or inspection commands that were actually run and their results.
- Any limitations that prevented full verification.
- The final Git status and confirmation of whether changes were confined to `port-blueprint/`.

Do not implement any part of the Go application in this session.

**Definition of success:** A new AI coding-agent session can open this repository, read `port-blueprint/`, understand the existing application and its behavior, follow the documented technical decisions, and begin a structured Go migration with minimal rediscovery, minimal guesswork, and a clear way to verify correctness.