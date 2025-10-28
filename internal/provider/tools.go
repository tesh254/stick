package provider

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    functions "github.com/tesh254/stick/internal/functions"
)

func (ts *Tools) MetadataLabels(name string) map[string]ToolMetadata {
    // Shell tools
    // RunCommand executes a shell command in the user's environment and returns combined stdout/stderr.
    // Useful for build/test/lint or inspecting environment state.
    runCommandMeta := ToolMetadata{
        Label:    "RunCommand",
        Info:     "Execute a shell command in the developer environment and return output.",
        Category: "Shell",
        Params: []ToolParam{
            {Name: "command", Type: "command", Description: "The shell command to execute.", Required: true},
            {Name: "cwd", Type: "path", Description: "Working directory to run the command in.", Required: false},
            {Name: "timeout_seconds", Type: "number", Description: "Optional timeout for command execution.", Required: false},
        },
        Returns: ToolReturn{Type: "string", Description: "Combined stdout/stderr output from the command."},
        Actions: []string{
            "spawn interactive or non-interactive shell command",
            "capture and return command output",
        },
        Examples: []string{"run_command: \"go test ./...\""},
    }

    // Filesystem tools
    // ReadFile reads the contents of a file by path.
    readFileMeta := ToolMetadata{
        Label:    "ReadFile",
        Info:     "Read the contents of a file from the filesystem.",
        Category: "Filesystem",
        Params: []ToolParam{
            {Name: "path", Type: "path", Description: "Absolute or project-relative file path.", Required: true},
            {Name: "encoding", Type: "string", Description: "Text encoding (default utf-8).", Required: false},
        },
        Returns: ToolReturn{Type: "string", Description: "File contents as a string."},
        Actions: []string{"open file and read contents"},
        Examples: []string{"read_file: {path: \"README.md\"}"},
    }

    // ReadDir lists entries in a directory.
    readDirMeta := ToolMetadata{
        Label:    "ReadDir",
        Info:     "List files and subdirectories for a given path.",
        Category: "Filesystem",
        Params: []ToolParam{
            {Name: "path", Type: "path", Description: "Directory path to list.", Required: true},
            {Name: "depth", Type: "number", Description: "Optional recursion depth (0 for non-recursive).", Required: false},
        },
        Returns: ToolReturn{Type: "array", Description: "List of entries (names or structured metadata)."},
        Actions: []string{"enumerate directory entries"},
        Examples: []string{"read_dir: {path: \"internal/\"}"},
    }

    // GetDirInfo returns structured metadata about a directory (sizes, types, modified times).
    getDirInfoMeta := ToolMetadata{
        Label:    "GetDirInfo",
        Info:     "Return structured metadata about a directory's contents.",
        Category: "Filesystem",
        Params: []ToolParam{{Name: "path", Type: "path", Description: "Directory to inspect.", Required: true}},
        Returns: ToolReturn{Type: "object", Description: "Structured JSON-like object describing directory entries."},
        Actions: []string{"stat directory and child entries"},
        Examples: []string{"get_dir_info: {path: \"internal/provider\"}"},
    }

    // Project tools
    // DetectVCS detects whether a directory is a version-controlled repository (e.g., git).
    detectVCSMeta := ToolMetadata{
        Label:    "DetectVCS",
        Info:     "Detect if a directory is under version control (e.g., Git).",
        Category: "Project",
        Params:   []ToolParam{{Name: "path", Type: "path", Description: "Directory to check (default CWD).", Required: false}},
        Returns:  ToolReturn{Type: "object", Description: "Object with fields like {system: 'git', root: '/path', detected: true}."},
        Actions:  []string{"inspect directory for VCS markers"},
        Examples: []string{"detect_vcs: {}"},
    }

    // ProjectDetection infers the project type/framework/language from files present.
    projectDetectionMeta := ToolMetadata{
        Label:    "ProjectDetection",
        Info:     "Infer project type (language/framework) from repository contents.",
        Category: "Project",
        Params:   []ToolParam{{Name: "path", Type: "path", Description: "Directory to analyze (default CWD).", Required: false}},
        Returns:  ToolReturn{Type: "object", Description: "Object describing detected languages, frameworks, and build tooling."},
        Actions:  []string{"scan files and config to infer project characteristics"},
        Examples: []string{"project_detection: {}"},
    }

    // WriteDir ensures a directory structure exists (create nested folders as needed).
    writeDirMeta := ToolMetadata{
        Label:    "WriteDir",
        Info:     "Create or ensure directory structure exists at the provided path.",
        Category: "Filesystem",
        Params: []ToolParam{
            {Name: "path", Type: "path", Description: "Directory path to create or ensure.", Required: true},
            {Name: "mode", Type: "string", Description: "Permissions/mode hint (platform-dependent).", Required: false},
        },
        Returns: ToolReturn{Type: "string", Description: "Status message indicating success or details."},
        Actions: []string{"create directory tree if missing"},
        Examples: []string{"write_dir: {path: \"internal/newpkg\"}"},
    }

    // WriteFile writes text content to a file (create or overwrite).
    writeFileMeta := ToolMetadata{
        Label:    "WriteFile",
        Info:     "Write content to a file, creating it if it does not exist.",
        Category: "Filesystem",
        Params: []ToolParam{
            {Name: "path", Type: "path", Description: "Destination file path.", Required: true},
            {Name: "content", Type: "string", Description: "File content to write.", Required: true},
            {Name: "overwrite", Type: "boolean", Description: "Allow overwriting existing files (default true).", Required: false},
            {Name: "encoding", Type: "string", Description: "Text encoding (default utf-8).", Required: false},
        },
        Returns: ToolReturn{Type: "string", Description: "Status or location confirmation."},
        Actions: []string{"write content to file path"},
        Examples: []string{"write_file: {path: \"README.md\", content: \"Hello\"}"},
    }

    // UpsertFile applies edits or writes content to ensure a file reaches a desired state.
    upsertFileMeta := ToolMetadata{
        Label:    "UpsertFile",
        Info:     "Create or update a file via content or structured edits.",
        Category: "Filesystem",
        Params: []ToolParam{
            {Name: "path", Type: "path", Description: "Target file.", Required: true},
            {Name: "content", Type: "string", Description: "Desired final content (used if edits not provided).", Required: false},
            {Name: "edits", Type: "array", Description: "List of patch operations or line edits.", Required: false},
            {Name: "strategy", Type: "string", Description: "Edit strategy (e.g., replace, merge).", Required: false},
        },
        Returns: ToolReturn{Type: "object", Description: "Object with resulting state (e.g., {changed: true})."},
        Actions: []string{"apply edits or replace file content"},
        Examples: []string{"upsert_file: {path: \"main.go\", edits: [...] }"},
    }

    // UpsertDir creates/updates a directory structure to match a target layout.
    upsertDirMeta := ToolMetadata{
        Label:    "UpsertDir",
        Info:     "Create or update a directory structure to match a target layout.",
        Category: "Filesystem",
        Params: []ToolParam{
            {Name: "path", Type: "path", Description: "Target directory.", Required: true},
            {Name: "structure", Type: "object", Description: "Desired directory tree description.", Required: true},
        },
        Returns: ToolReturn{Type: "object", Description: "Object describing operations performed (created/updated)."},
        Actions: []string{"create missing directories and update structure"},
        Examples: []string{"upsert_dir: {path: \"internal\", structure: {...}}"},
    }

    // Web tools
    // WebSearch queries the web and returns results summary.
    webSearchMeta := ToolMetadata{
        Label:    "WebSearch",
        Info:     "Perform a web search with a query and optional result limit.",
        Category: "Web",
        Params: []ToolParam{
            {Name: "query", Type: "string", Description: "Search terms.", Required: true},
            {Name: "top_k", Type: "number", Description: "Max results to return.", Required: false},
        },
        Returns: ToolReturn{Type: "array", Description: "List of search results (titles/links/summaries)."},
        Actions: []string{"query web search provider and summarize results"},
        Examples: []string{"web_search: {query: \"Go generics examples\", top_k: 5}"},
    }

    // Function tools
    // CallStdFunc calls a built-in Stick function by name using the functions registry.
    callStdFuncMeta := ToolMetadata{
        Label:    "CallStdFunc",
        Info:     "Call a built-in Stick standard function by name with arguments.",
        Category: "Functions",
        Params: []ToolParam{
            {Name: "function", Type: "string", Description: "Function name to call.", Required: true},
            {Name: "args", Type: "array", Description: "Positional string arguments for the function.", Required: false},
        },
        Returns: ToolReturn{Type: "object", Description: "Function result and optional metadata."},
        Actions: []string{"route to functions registry and execute"},
        Examples: []string{"call_std_func: {function: \"echo\", args: [\"hello\"]}"},
    }

    meta := map[string]ToolMetadata{
        // Shell
        "run_command": runCommandMeta,
        // Filesystem
        "read_file":   readFileMeta,
        "read_dir":    readDirMeta,
        "get_dir_info": getDirInfoMeta,
        "write_dir":   writeDirMeta,
        "write_file":  writeFileMeta,
        "upsert_file": upsertFileMeta,
        "upsert_dir":  upsertDirMeta,
        // Project
        "detect_vcs":        detectVCSMeta,
        "project_detection": projectDetectionMeta,
        // Web
        "web_search": webSearchMeta,
        // Functions
        "call_std_func": callStdFuncMeta,
    }

    // Optional validation; returns map regardless but ensures definitions are internally consistent.
    _ = ValidateToolMetadataMap(meta)
    return meta
}

// ValidateToolMetadataMap checks that labels, categories, params, and return types are consistent.
func ValidateToolMetadataMap(meta map[string]ToolMetadata) error {
    allowedParamTypes := map[string]bool{
        "string": true, "number": true, "boolean": true, "array": true, "object": true,
        "path": true, "command": true, "json": true,
    }
    allowedReturnTypes := map[string]bool{"string": true, "number": true, "boolean": true, "array": true, "object": true}
    allowedCategories := map[string]bool{"Shell": true, "Filesystem": true, "Web": true, "Project": true, "Functions": true}
    labelRe := regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)

    for key, tm := range meta {
        if key == "" {
            return fmt.Errorf("tool key cannot be empty")
        }
        if tm.Label == "" || !labelRe.MatchString(tm.Label) {
            return fmt.Errorf("invalid label for %s: %q", key, tm.Label)
        }
        if tm.Info == "" {
            return fmt.Errorf("info is required for %s", key)
        }
        if tm.Category == "" || !allowedCategories[tm.Category] {
            return fmt.Errorf("invalid category for %s: %q", key, tm.Category)
        }
        if !allowedReturnTypes[tm.Returns.Type] {
            return fmt.Errorf("invalid return type for %s: %q", key, tm.Returns.Type)
        }
        // Params validation
        seen := map[string]bool{}
        for _, p := range tm.Params {
            if p.Name == "" {
                return fmt.Errorf("param name is required for %s", key)
            }
            if seen[p.Name] {
                return fmt.Errorf("duplicate param %q for %s", p.Name, key)
            }
            seen[p.Name] = true
            if !allowedParamTypes[p.Type] {
                return fmt.Errorf("invalid param type %q for %s.%s", p.Type, key, p.Name)
            }
        }
        if len(tm.Actions) == 0 {
            return fmt.Errorf("at least one action is required for %s", key)
        }
    }
    return nil
}

// ---- Channel output formatting helpers ----

// toolOutputPayload defines the structured content we embed in a single string
// for agent model consumption. The final channel emits strings containing a
// JSON payload prefixed with a marker for easy routing in TUI.
type toolOutputPayload struct {
    Tool     string            `json:"tool"`
    Status   string            `json:"status"` // success|error
    Data     string            `json:"data"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

// formatToolOutput converts payload to a single string for the agent model.
// The output is a line prefixed with [TOOL_OUTPUT] followed by a JSON object.
func formatToolOutput(p toolOutputPayload) string {
    b, _ := json.Marshal(p)
    return "[TOOL_OUTPUT] " + string(b)
}

// makeMetadata helper to avoid nil maps.
func makeMetadata(kv ...string) map[string]string {
    m := map[string]string{}
    for i := 0; i+1 < len(kv); i += 2 {
        m[kv[i]] = kv[i+1]
    }
    return m
}

// ---- Common helpers ----

// getDefaultShell returns a best-effort default shell for command execution.
func getDefaultShell() string {
    if shell := os.Getenv("SHELL"); shell != "" {
        return shell
    }
    return "/bin/sh"
}

// ---- Tool argument structs ----

type RunCommandArgs struct {
    Command        string
    CWD            string
    TimeoutSeconds int
}

type ReadFileArgs struct {
    Path     string
    Encoding string // informational; we treat content as utf-8 text
}

type ReadDirArgs struct {
    Path  string
    Depth int
}

type GetDirInfoArgs struct {
    Path string
}

type DetectVCSArgs struct {
    Path string // default cwd if empty
}

type ProjectDetectionArgs struct {
    Path string // default cwd if empty
}

type WriteDirArgs struct {
    Path string
    Mode string // informational only
}

type WriteFileArgs struct {
    Path      string
    Content   string
    Overwrite bool
    Encoding  string // informational only
}

type UpsertFileArgs struct {
    Path     string
    Content  string
    Edits    []string // placeholder for future structured edits
    Strategy string
}

type UpsertDirArgs struct {
    Path      string
    Structure map[string]interface{} // placeholder for future layout description
}

type WebSearchArgs struct {
    Query string
    TopK  int
}

type CallStdFuncArgs struct {
    Function string
    Args     []string
}

// ---- Individual local tool functions (channel-based) ----

// runCommand executes a shell command and emits a single formatted string on a channel.
// Purpose: Run a developer command (build/test/lint/etc.).
// Usage: Provide Command string; optional CWD and TimeoutSeconds.
// Channel output format: [TOOL_OUTPUT] {tool, status, data, metadata}
func runCommand(args RunCommandArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    shell := getDefaultShell()

    // prepare command context with optional timeout
    var ctx context.Context
    var cancel context.CancelFunc
    if args.TimeoutSeconds > 0 {
        ctx, cancel = context.WithTimeout(context.Background(), time.Duration(args.TimeoutSeconds)*time.Second)
    } else {
        ctx, cancel = context.WithCancel(context.Background())
    }

    go func() {
        defer cancel()
        defer close(ch)

        cmd := exec.CommandContext(ctx, shell, "-c", args.Command)
        if strings.TrimSpace(args.CWD) != "" {
            cmd.Dir = args.CWD
        }

        // Capture combined output
        out, err := cmd.CombinedOutput()
        data := string(out)
        meta := makeMetadata("cwd", cmd.Dir, "shell", shell)
        if ctx.Err() == context.DeadlineExceeded {
            ch <- formatToolOutput(toolOutputPayload{Tool: "run_command", Status: "error", Data: "timeout exceeded", Metadata: meta})
            return
        }
        if err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "run_command", Status: "error", Data: data + "\n" + err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "run_command", Status: "success", Data: data, Metadata: meta})
    }()

    // Return a final formatted string as immediate context (empty; will be same as channel payload once received).
    // For pre-integration, we return a placeholder noting streaming via channel.
    final := formatToolOutput(toolOutputPayload{Tool: "run_command", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// readFile reads file contents and emits result via channel.
// Purpose: Provide file content to the agent for context.
// Params: ReadFileArgs{Path string, Encoding string}
// Channel output format: [TOOL_OUTPUT] {tool:"read_file", status, data, metadata:{path,encoding}}
func readFile(args ReadFileArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        b, err := os.ReadFile(args.Path)
        meta := makeMetadata("path", args.Path, "encoding", defaultString(args.Encoding, "utf-8"))
        if err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "read_file", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "read_file", Status: "success", Data: string(b), Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "read_file", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// readDir lists directory entries and emits a JSON array string of names.
// Purpose: Quickly enumerate directory entries for navigation/context.
// Params: ReadDirArgs{Path string, Depth int}
// Channel output format: [TOOL_OUTPUT] {tool:"read_dir", status, data:{names, count}, metadata:{path, depth}}
func readDir(args ReadDirArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        entries, err := os.ReadDir(args.Path)
        meta := makeMetadata("path", args.Path, "depth", fmt.Sprintf("%d", args.Depth))
        if err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "read_dir", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        names := make([]string, 0, len(entries))
        for _, e := range entries {
            names = append(names, e.Name())
        }
        payload := struct {
            Names []string `json:"names"`
            Count int      `json:"count"`
        }{Names: names, Count: len(names)}
        dataBytes, _ := json.Marshal(payload)
        ch <- formatToolOutput(toolOutputPayload{Tool: "read_dir", Status: "success", Data: string(dataBytes), Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "read_dir", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// getDirInfo returns structured metadata for directory contents (type/size/modtime minimal).
// Purpose: Provide richer directory context (sizes, types, mod times).
// Params: GetDirInfoArgs{Path string}
// Channel output format: [TOOL_OUTPUT] {tool:"get_dir_info", status, data:{entries, count}, metadata:{path}}
func getDirInfo(args GetDirInfoArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        entries, err := os.ReadDir(args.Path)
        meta := makeMetadata("path", args.Path)
        if err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "get_dir_info", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        type entryInfo struct {
            Name    string    `json:"name"`
            IsDir   bool      `json:"is_dir"`
            Size    int64     `json:"size"`
            ModTime time.Time `json:"mod_time"`
        }
        infos := make([]entryInfo, 0, len(entries))
        for _, e := range entries {
            fi, err := os.Stat(filepath.Join(args.Path, e.Name()))
            if err != nil {
                continue
            }
            infos = append(infos, entryInfo{Name: e.Name(), IsDir: e.IsDir(), Size: fi.Size(), ModTime: fi.ModTime()})
        }
        payload := struct {
            Entries []entryInfo `json:"entries"`
            Count   int         `json:"count"`
        }{Entries: infos, Count: len(infos)}
        dataBytes, _ := json.Marshal(payload)
        ch <- formatToolOutput(toolOutputPayload{Tool: "get_dir_info", Status: "success", Data: string(dataBytes), Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "get_dir_info", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// detectVCS checks for common VCS markers in a directory (e.g., .git).
// Purpose: Determine if repository has version control and basic info.
// Params: DetectVCSArgs{Path string}
// Channel output format: [TOOL_OUTPUT] {tool:"detect_vcs", status, data:{system,root,detected}, metadata:{path}}
func detectVCS(args DetectVCSArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        p := args.Path
        if strings.TrimSpace(p) == "" {
            p = "."
        }
        gitPath := filepath.Join(p, ".git")
        _, err := os.Stat(gitPath)
        meta := makeMetadata("path", p)
        payload := struct {
            System   string `json:"system"`
            Root     string `json:"root"`
            Detected bool   `json:"detected"`
        }{System: "git", Root: p}
        if err == nil {
            payload.Detected = true
        } else if errors.Is(err, os.ErrNotExist) {
            payload.Detected = false
        } else {
            ch <- formatToolOutput(toolOutputPayload{Tool: "detect_vcs", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        b, _ := json.Marshal(payload)
        ch <- formatToolOutput(toolOutputPayload{Tool: "detect_vcs", Status: "success", Data: string(b), Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "detect_vcs", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// projectDetection infers project details from common files.
// Purpose: Identify languages/frameworks to tailor agent actions.
// Params: ProjectDetectionArgs{Path string}
// Channel output format: [TOOL_OUTPUT] {tool:"project_detection", status, data:{...detected...}, metadata:{path}}
func projectDetection(args ProjectDetectionArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        p := args.Path
        if strings.TrimSpace(p) == "" {
            p = "."
        }
        meta := makeMetadata("path", p)
        detected := map[string]bool{}
        if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
            detected["go"] = true
        }
        if _, err := os.Stat(filepath.Join(p, "package.json")); err == nil {
            detected["node"] = true
        }
        if _, err := os.Stat(filepath.Join(p, "Cargo.toml")); err == nil {
            detected["rust"] = true
        }
        b, _ := json.Marshal(detected)
        ch <- formatToolOutput(toolOutputPayload{Tool: "project_detection", Status: "success", Data: string(b), Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "project_detection", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// writeDir ensures a directory exists.
// Purpose: Create necessary directories prior to writing files.
// Params: WriteDirArgs{Path string, Mode string}
// Channel output format: [TOOL_OUTPUT] {tool:"write_dir", status, data:"created"|error, metadata:{path,mode}}
func writeDir(args WriteDirArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        err := os.MkdirAll(args.Path, 0o755)
        meta := makeMetadata("path", args.Path, "mode", defaultString(args.Mode, "0755"))
        if err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "write_dir", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "write_dir", Status: "success", Data: "created", Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "write_dir", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// writeFile writes content to a file, optionally respecting overwrite flag.
// Purpose: Persist agent-generated content to disk.
// Params: WriteFileArgs{Path string, Content string, Overwrite bool, Encoding string}
// Channel output format: [TOOL_OUTPUT] {tool:"write_file", status, data:"written"|error, metadata:{path,overwrite,encoding}}
func writeFile(args WriteFileArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        meta := makeMetadata("path", args.Path, "overwrite", fmt.Sprintf("%t", defaultBool(args.Overwrite, true)), "encoding", defaultString(args.Encoding, "utf-8"))
        if _, err := os.Stat(args.Path); err == nil && !args.Overwrite {
            ch <- formatToolOutput(toolOutputPayload{Tool: "write_file", Status: "error", Data: "file exists and overwrite=false", Metadata: meta})
            return
        }
        if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "write_file", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "write_file", Status: "success", Data: "written", Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "write_file", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// upsertFile creates or updates a file. Edits are currently unsupported and will return an error.
// Purpose: Reach a desired file state via content or edits.
// Params: UpsertFileArgs{Path string, Content string, Edits []string, Strategy string}
// Channel output format: [TOOL_OUTPUT] {tool:"upsert_file", status, data:"upserted"|error, metadata:{path,strategy}}
func upsertFile(args UpsertFileArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        meta := makeMetadata("path", args.Path, "strategy", defaultString(args.Strategy, "replace"))
        if len(args.Edits) > 0 {
            ch <- formatToolOutput(toolOutputPayload{Tool: "upsert_file", Status: "error", Data: "structured edits unsupported in pre-integration", Metadata: meta})
            return
        }
        if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "upsert_file", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "upsert_file", Status: "success", Data: "upserted", Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "upsert_file", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// upsertDir ensures the directory exists; structure handling is deferred to integration.
// Purpose: Create/update directory structure to match a target layout.
// Params: UpsertDirArgs{Path string, Structure map[string]interface{}}
// Channel output format: [TOOL_OUTPUT] {tool:"upsert_dir", status, data:"upserted"|error, metadata:{path}}
func upsertDir(args UpsertDirArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        meta := makeMetadata("path", args.Path)
        if err := os.MkdirAll(args.Path, 0o755); err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "upsert_dir", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "upsert_dir", Status: "success", Data: "upserted", Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "upsert_dir", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// webSearch is a stub for pre-integration; it emits an error indicating not implemented.
// Purpose: Retrieve web results when integrated with a provider.
// Params: WebSearchArgs{Query string, TopK int}
// Channel output format: [TOOL_OUTPUT] {tool:"web_search", status:"error", data:"web search not implemented", metadata:{query,top_k}}
func webSearch(args WebSearchArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        meta := makeMetadata("query", args.Query, "top_k", fmt.Sprintf("%d", args.TopK))
        ch <- formatToolOutput(toolOutputPayload{Tool: "web_search", Status: "error", Data: "web search not implemented", Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "web_search", Status: "error", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// callStdFunc calls a registered standard function from the functions registry.
// Purpose: Execute a named function with positional args via registry.
// Params: CallStdFuncArgs{Function string, Args []string}
// Channel output format: [TOOL_OUTPUT] {tool:"call_std_func", status, data:<function result>|error, metadata:{function}}
func callStdFunc(args CallStdFuncArgs) (string, <-chan string) {
    ch := make(chan string, 1)
    go func() {
        defer close(ch)
        r := functions.NewRegistry()
        // Register a baseline set of built-ins; integration can expand.
        r.Register("add", functions.Add, 0, 2)
        r.Register("echo", functions.Echo, 0, -1)
        r.Register("get_llm_text", functions.GetLLMText, 1, 1)
        r.Register("get_page_html_content_to_markdown", functions.GetPageHTMLContentToMarkdown, 1, 1)

        res, err := r.Call(args.Function, args.Args)
        meta := makeMetadata("function", args.Function)
        if err != nil {
            ch <- formatToolOutput(toolOutputPayload{Tool: "call_std_func", Status: "error", Data: err.Error(), Metadata: meta})
            return
        }
        ch <- formatToolOutput(toolOutputPayload{Tool: "call_std_func", Status: "success", Data: res, Metadata: meta})
    }()
    final := formatToolOutput(toolOutputPayload{Tool: "call_std_func", Status: "success", Data: "", Metadata: makeMetadata("note", "use channel output")})
    return final, ch
}

// ---- Small utilities ----

func defaultString(s, def string) string {
    if strings.TrimSpace(s) == "" {
        return def
    }
    return s
}

func defaultBool(b, def bool) bool {
    // if b is false but set explicitly we still return b; here we simply return b
    // because we cannot detect unset boolean without pointer.
    return b
}
