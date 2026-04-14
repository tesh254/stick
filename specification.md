# Stick Project Specification

## 1. Progress Summary
The Stick project has evolved from a simple command-line tool for Git repositories to a more sophisticated AI-enhanced application with a terminal user interface (TUI) and various backend components. Key completed milestones and features include:

### Completed Milestones
- **Initial Setup and Core Functionality**: 
  - Project initialization with Go 1.24+.
  - Basic commands for Git operations (e.g., show commit diff, version info).
  - Installation script and build process established.

- **Database Integration**:
  - Implementation of a SQLite-based database for storing conversations, messages, call events, and usage data.
  - Repository patterns for CRUD operations on conversations, messages, calls, and usage.

- **Function Registry and Parser**:
  - A registry for managing functions with support for min/max arguments.
  - Parser for handling function calls in inputs.

- **TUI Development**:
  - Built using Bubble Tea framework.
  - Core TUI components: viewport for message display, textarea for input, slash command modal.
  - Message handling, command history, and function call processing.
  - Persistence of messages and call events to DB.

- **AI Agent Integration**:
  - Agent session management with tools and types.
  - Prompts system for AI interactions, including system prompts and short prompts.

- **Additional Components**:
  - Web crawling and markdown processing for RAG (Retrieval-Augmented Generation).
  - Event emitter for system events.
  - Studio API server with routes, handlers, and configuration.
  - Utility functions for user management and version/build info.

- **Recent Enhancements**:
  - Examination of existing TUI structure.
  - Design and implementation of a yes/no list component (`ask_user.go`) using Bubble Tea.
  - Partial integration of `askUser` mode into the TUI model, including new fields for mode tracking.

### Completed Features
- Git-related commands (commit diff, version).
- Persistent storage for conversations and function calls.
- Interactive TUI with slash commands (/functions, /help_, /reload, /replay_last).
- Function execution and rendering in TUI.
- AI provider client with tool support.
- RAG embeddings for enhanced querying.

## 2. Current Technical Specifications and Architecture

### Technical Specifications
- **Language**: Go 1.24+
- **Dependencies**: Managed via `go.mod` (includes Bubble Tea, Lipgloss, SQLite, etc.).
- **Database**: SQLite with custom repositories for entities like Message, Conversation, CallEvent, Usage.
- **TUI Framework**: Bubble Tea for interactive terminal UI.
- **AI Integration**: Custom agent with prompts and tools; RAG for context-aware responses.
- **API**: Studio server providing endpoints for configuration and interactions.

### Architecture Diagrams
#### High-Level Architecture
```
+-------------------+     +-------------------+
|     CLI/TUI       |     |    Studio API     |
| (Bubble Tea)      |<--->| (HTTP Server)     |
+-------------------+     +-------------------+
            |                        |
            v                        v
    +----------------+     +-----------------+
    | AI Agent/Tools |     | Web Crawler/RAG |
    +----------------+     +-----------------+
            |                        |
            v                        v
    +----------------+     +-----------------+
    |   Database     |     | File System/FS  |
    |   (SQLite)     |     |   Utilities     |
    +----------------+     +-----------------+
```

#### TUI Component Diagram
```
+--------------------+
|    Viewport        |  (Displays messages)
+--------------------+
|    Slash Modal     |  (For commands like /functions)
+--------------------+
|    Textarea        |  (User input)
+--------------------+
| AskUser Mode (New) |  (Yes/No confirmation list)
+--------------------+
```

## 3. Changes from Original Project Scope
- **Original Scope** (from README): A simple command-line tool for Git repositories, providing commit info, branches, and versions.
- **Changes**:
  - Expanded to include a full TUI for interactive sessions.
  - Integrated AI agent capabilities for intelligent interactions, including tool calls and bash command confirmations.
  - Added database persistence, web crawling, RAG, and a web API (Studio).
  - Introduction of `askUser` mode for user confirmations in AI workflows, deviating from pure Git focus to AI-assisted development tools.
  - Scope broadened to support AI-driven tasks like installations via bash, with user approvals.

These changes transform Stick into an AI-powered development assistant rather than a basic Git utility.

## 4. Known Issues and Outstanding Tasks
- **Known Issues**:
  - Potential inconsistencies in message persistence during high-volume interactions (addressed partially with non-blocking writes).
  - Limited error handling in function parsing for complex inputs.
  - TUI resizing may not always adjust modal heights perfectly.
  - No comprehensive testing for RAG embeddings in production scenarios.

- **Outstanding Tasks** (from ongoing development):
  - Fully integrate the yes/no list as a new mode in the TUI model, toggling the textarea appropriately.
  - Hook up the new `askUser` mode to be triggered by AI agent tool calls, especially for bash command confirmations like installations.
  - Test the `askUser` implementation to ensure it works as expected.
  - Enhance documentation in sub-modules (e.g., agent, crawl).
  - Implement authentication for Studio API.

## 5. Upcoming Development Priorities and Timelines
- **Short-term (Next 1-2 weeks)**:
  - Complete `askUser` mode integration and testing.
  - Add unit tests for TUI components and AI agent tools.

- **Medium-term (Next 1-2 months)**:
  - Expand RAG capabilities with better embedding models.
  - Enhance Studio API with more endpoints and security features.
  - Optimize database migrations and add indexing for performance.

- **Long-term (3+ months)**:
  - Integrate more AI providers and advanced prompting techniques.
  - Develop web-based UI complementing the TUI.
  - Community contributions: Improve contributing guidelines and add examples.

This specification serves as a complete reference. For code-level details, refer to inline comments and module-specific docs.