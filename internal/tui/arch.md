# TUI Message Architecture

This module implements a dual-slice architecture for chat messages:

- Display slice (`model.messages []string`): optimized for real-time rendering.
  - Contains styled strings produced by the TUI, including prefixes and function call blocks.
  - Updated synchronously to ensure immediate UI feedback and correct scroll behavior.

- Storage slice (`model.storageMessages []*db.Message`): structured for persistence.
  - Captures canonical message data: `id`, `conversation`, `content`, `role`, `created_at`.
  - Updated in lockstep with the display slice to preserve chronological integrity.

## Transformation Pipeline

- Raw input is processed into both display and storage representations.
- For function calls, the pipeline produces:
  - Styled display output via `FunctionRenderer` (name + result or error block).
  - Raw result (or error text) for storage as an assistant message.
- Slash commands produce a UI response that is also stored as an assistant message.

## Synchronization and Performance

- Storage operations use a background worker and a buffered queue (`model.storageQueue`) to avoid blocking the UI thread.
- The worker persists messages using the repository manager with simple retries for transient errors.
- A lightweight validator checks display/storage slice counts to catch inconsistencies without impacting performance.

## Chronological Order and Integrity

- Both slices are updated synchronously (append order) at the time of message creation.
- Persistence happens asynchronously; failures are logged and do not block rendering.

## Testing Guidance

- Verify transformation correctness across user input, slash commands, and function calls.
- Validate display/storage consistency under high message volume by enqueuing many messages and ensuring the slices stay in sync.