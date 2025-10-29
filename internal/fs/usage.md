# File Updater Package Usage

This document describes the usage of the public methods in the `fileupdater` package.

## `UpdateFileWithDiff`

The main function for computing and optionally applying structured, line-based diffs to files.

### Signature
```go
func UpdateFileWithDiff(originalPath, newContent string, contextLines int, apply bool) (FileDiff, bool, error)
```

### Parameters
- `originalPath`: Path to the file to update
- `newContent`: Proposed full new content as string
- `contextLines`: Number of unchanged lines to show around changes (e.g., 3 like Git)
- `apply`: If true, apply the changes atomically

### Returns
- `FileDiff`: Structured diff representation
- `bool`: Whether the changes were applied
- `error`: Error if any occurred during the process

### Description
`UpdateFileWithDiff` computes a structured, line-based diff using the Longest Common Subsequence (LCS) algorithm and optionally applies it atomically. If the file content has not changed, it returns early with an empty diff.

When `apply` is true, the function performs an atomic write operation by:
1. Verifying the diff application matches the expected result
2. Creating a temporary file with the new content
3. Renaming the temporary file to the original path

### Example Usage
```go
package main

import (
    "fmt"
    fileupdater "github.com/tesh254/stick/internal/fs"
)

func main() {
    // Path to original file and proposed new content.
    path := "./example.go"
    proposed := "package main\n\nfunc main() {\n    println(\"hello\")\n}\n"
    
    // Compute diff without applying.
    diff, applied, err := fileupdater.UpdateFileWithDiff(path, proposed, 3, false)
    if err != nil {
        panic(err)
    }
    fmt.Println("Applied:", applied)
    fmt.Println(fileupdater.FormatDiffForTerminal(diff))
    
    // Apply changes atomically.
    diff, applied, err = fileupdater.UpdateFileWithDiff(path, proposed, 3, true)
    if err != nil {
        panic(err)
    }
    fmt.Println("Applied:", applied)
}
```

## `FormatDiffForTerminal`

Formats a FileDiff struct into a unified-diff style string suitable for CLI output.

### Signature
```go
func FormatDiffForTerminal(fd FileDiff) string
```

### Parameters
- `fd`: The FileDiff to format

### Returns
- `string`: Unified-diff style formatted string

### Description
This function generates a string representation of the diff in unified diff format with headers that show the line numbers and counts for both original and new content.

### Example Output
```
@@ -startOrig,countOrig +startNew,countNew @@
 context line
-removed line
+added line
```