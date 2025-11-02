# TUI Package Breakdown

This document provides a practical, developer‑focused overview of how the TUI works, how data flows through it, and where to extend functionality. If you’re onboarding to the TUI, start here.

## Goals and Design
- Lightweight chat interface built with Bubble Tea (`bubbletea`), `bubbles` (`textarea`, `viewport`), and `lipgloss` for styling.
- Dual representation of chat data:
  - Display slice: `model.messages []string` for styled, fast rendering.
  - Storage slice: `model.storageMessages []*db.Message` for canonical persistence.
- Non‑blocking persistence via background workers and buffered queues.
- Consistent, styled function/tool call rendering and metadata storage.
- Easy reconstruction of conversations from storage (including recorded tool calls).

See `arch.md` for a visual summary of the dual‑slice architecture.

## Key Files and Responsibilities
- `model.go`: Defines the core `model` struct and `NewProgram()` entry point.
- `initial.go`: Initializes UI components, function registry, DB and conversation; starts persistence workers.
- `update_view.go`: Bubble Tea `Update` and `View` loop, slash modal rendering, viewport management.
- `messages.go`: Input handling, history navigation, slash commands, function call pipeline, storage linking.
- `message_pipeline.go`: Function call parsing + execution producing both styled display and raw result.
- `ui.go`: `FunctionRenderer` with unified API for styled function blocks.
- `reconstruction.go`: Rebuilds the display from stored messages and recorded call events; supports replay.
- `storage_sync.go`: Non‑blocking persistence workers for messages and call events; consistency checks.
- `styles.go`: Shared styles for headers, blocks, and function/tool result formatting.
- `renders.go`: Legacy/simple render helpers (most new code uses `FunctionRenderer`).
- `ascii.go`: Startup banner (`AGENT_ASCII`) used in the initial viewport content.
- `db/list_conversations.go`: Separate mini‑TUI (package `dbtui`) for listing conversations.

## Lifecycle Overview
- Program startup:
  - `handlers.StartSession()` calls `tui.NewProgram()` and runs it.
  - `initial.go/initialModel()` builds the initial `model`:
    - `setupTextarea()` and `setupViewport()` configure input + main view.
    - `setupFunctionRegistry()` registers built‑in functions.
    - `setupDBAndConversation()` opens local DB, creates a new conversation, initializes queues.
  - `model.Init()` starts background workers:
    - `startStorageWorker(&m)` for messages.
    - `startCallStorageWorker(&m)` for function call events.

- Update + View loop (`update_view.go`):
  - `Update(msg tea.Msg)` updates textarea and viewport; tracks slash mode; delegates key handling.
  - `View()` renders viewport + optional slash search modal + textarea.

## Model Structure (selected)
- Display: `messages []string` — styled lines for viewport rendering.
- Storage: `storageMessages []*db.Message` — canonical entries destined for DB.
- Inputs/UI: `textarea`, `viewport`, `wrapStyle`, `senderStyle`, `viewportFocused`.
- History: `commandHistory []string`, `historyIndex int`.
- Registry: `functionRegistry *functions.Registry`.
- Persistence: `dbConn *db.DB`, `repoManager db.RepositoryManager`, `conversationID uuidv7.UUID`.
- Queues: `storageQueue chan *db.Message`, `callStorageQueue chan *db.CallEvent`.
- Slash modal: `showSearchModal`, `searchInput`, `allSlashCommands`, `filteredCommands`, `selectedIndex`, `isInSlashMode`.

## Input Handling and Flow (`messages.go`)
- `handleKeyMsg(...)` manages:
  - Viewport focus and scrolling keys (Up/Down/PageUp/PageDown/Home/End).
  - Global keys (Ctrl‑C, Esc to quit or exit slash mode, Ctrl‑F to focus viewport).
  - Enter → `handleEnterKey()` submission.
  - Up/Down → command history navigation when viewport isn’t focused.

- `handleEnterKey()`:
  - Captures input and appends a styled user line (with `{"username"}: ` prefix).
  - Creates/stores a `db.Message` for the user input; saves parent reference.
  - If input is a slash command → `processSlashCommand()` handles and stores a linked assistant message.
  - Else → attempts `processFunctionCallPipeline(input)`:
    - Produces `PipelineResult` with both `Display` (styled) and `RawResult` (storage).
    - Appends styled result to `messages`.
    - Stores a linked assistant message (`ParentMessage` set to the user message ID).
    - Builds and enqueues a `db.CallEvent` with timings and success/error metadata.
  - Updates viewport content and scroll position.

## Slash Commands and Modal
- Modal UI (auto‑starts when typing `/` at start of input):
  - `startSlashMode()`, `updateFilteredCommands()`, `renderSearchModal()`, `handleSearchModalKeys()`.
  - Shows a filterable list of available commands; Enter selects and closes modal.

- Supported commands (`processSlashCommand`):
  - `/functions` — lists registered functions.
  - `/help_{function}` — shows arg constraints for a function.
  - `/reload` — reloads and reconstructs current conversation from DB.
  - `/replay_last` — replays the most recent recorded function/tool call.

## Function Call Pipeline (`message_pipeline.go` + `ui.go`)
- Detection rules:
  - Must include `(` with no whitespace before it (e.g., `add(1, 2)`); single words or spaced tokens are treated as plain text.
- Parsing: `functions.Parser.ParseDetailed(...)` returns structure with `HasFunction`, `FunctionName`, `Arguments`, `Error`.
- Execution:
  - Raw execution: `functionRegistry.Call(name, args)` returns plain result for storage.
  - Styled rendering: `FunctionRenderer` combines a name header and result/error block.
- Output: `PipelineResult` includes `Display`, `RawResult`, timestamps (`StartedAt`, `CompletedAt`), and flags.
- Metadata: `buildCallEvent(...)` stores `db.CallEvent` with `ParentMessage`, status, timings, and `ParamsJSON`.

## Persistence and Sync (`storage_sync.go`)
- Messages:
  - `startStorageWorker(m)` consumes `storageQueue` and calls `repoManager.Messages().Create(...)` with basic retry.
  - `enqueueStorageMessage(msg)` pushes to queue without blocking.
- Call events:
  - `startCallStorageWorker(m)` consumes `callStorageQueue` and calls `repoManager.Calls().Create(...)` with basic retry.
  - `enqueueCallEvent(ev)` pushes to queue without blocking.
- Consistency:
  - `validateConsistency()` checks display vs storage slice counts (lightweight sanity).

## Reconstruction and Replay (`reconstruction.go`)
- `LoadConversationFromDB(conversationID)`:
  - Fetches messages via `Conversations().GetWithMessages(...)` and resets model buffers.
  - Fetches all `CallEvent` records and indexes them by `ParentMessage`.
  - Rebuilds the display chronologically:
    - Renders user lines with username prefix.
    - For each user message, replays associated calls from stored metadata first (styled name+result blocks).
    - Adds assistant messages while preventing duplicates when calls already rendered results.
  - Updates viewport and scroll.
- `ReplayCallFromEvent(ev)`:
  - Parses `ParamsJSON`, renders the name block, executes via registry, returns styled blocks for appending to the viewport.

## Styling (`styles.go`, `ui.go`, `renders.go`)
- Styles are centralized in `styles.go` (headers, body, error/result variants).
- `FunctionRenderer` in `ui.go` is the preferred API to render name and result/error blocks.
- `renders.go` hosts simple render helpers used by earlier code paths.

## Function Registry (`initial.go`, `internal/functions/*`)
- `setupFunctionRegistry()` registers core functions like:
  - `add`, `echo`, `print_statement`.
  - Crawl helpers like `get_llm_text`, `get_page_content`.
- See `internal/functions/` for implementation, parser details, and tests.

## Subpackage: DB TUI (`internal/tui/db/list_conversations.go`)
- Package name: `dbtui`.
- Provides a Bubble Tea list UI to browse conversations from the DB.
- Useful for administrative views; separate from the chat TUI flow.

## Testing
- `messages_pipeline_test.go`, `messages_test.go`, `messages_volume_test.go`, `ui_test.go` cover:
  - Dual‑slice consistency across user input and function calls.
  - Parser edge cases and styled output correctness.
  - High‑volume stability of display/storage synchronization.

## Extending the TUI
- Add a new function:
  - Implement in `internal/functions/`, register in `setupFunctionRegistry()`.
  - The parser will recognize `name(args...)` syntax; rendering uses `FunctionRenderer` automatically.
- Add a slash command:
  - Extend `processSlashCommand()` and include it in modal filtering via `populateSlashCommands()`.
- Persist additional metadata:
  - Augment `db.CallEvent` (schema + repository) and update `buildCallEvent()`.
- Change styling:
  - Edit `styles.go` and ensure `FunctionRenderer` and related helpers consume the updated styles.

## Common Gotchas
- Display/storage slice mismatch: check places that append to `messages` also enqueue a corresponding storage entry.
- Parser behavior: only `name(` starts function parsing; `name (` or `name` alone is treated as plain text.
- DB unavailable: the TUI still runs; storage queues and reconstruction gracefully no‑op.
- Viewport focus: when focused, arrow keys scroll history rather than command history.
- Slash mode vs input: the modal intercepts keys when active; closing it returns to normal input.

## Start a Session
- From the CLI, the app calls `handlers.StartSession()` which runs `tui.NewProgram()` in AltScreen mode.
- On start, you’ll see the ASCII banner from `ascii.go` and a prompt. Type a message or a function call like `add(2, 3)`, or use slash commands.