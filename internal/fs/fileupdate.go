package fileupdater

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Returns: Structured diff, whether applied, and error.
// UpdateFileWithDiff computes a structured, line-based diff and optionally applies it atomically.
// Input:
// - originalPath: Path to the file to update.
// - newContent: Proposed full new content as string.
// - contextLines: Number of unchanged lines to show around changes (e.g., 3 like Git).
// - apply: If true, apply the changes atomically.
// Output: Structured diff, whether applied, and error.
func UpdateFileWithDiff(originalPath, newContent string, contextLines int, apply bool) (FileDiff, bool, error) {
	originalBytes, err := os.ReadFile(originalPath)
	if err != nil {
		return FileDiff{}, false, errors.Wrap(err, "failed to read original file")
	}

	originalLines := splitLines(string(originalBytes))
	newLines := splitLines(newContent)
	originalHasTrailingNewline := strings.HasSuffix(string(originalBytes), "\n")
	newHasTrailingNewline := strings.HasSuffix(newContent, "\n")

	// If nothing changed, return early with empty diff.
	if linesEqual(originalLines, newLines) && originalHasTrailingNewline == newHasTrailingNewline {
		return FileDiff{Hunks: nil}, false, nil
	}

	// Compute LCS matrix diff
	lcs := computeLCS(originalLines, newLines)

	// Build hunks from LCS backtrace
	hunks := buildHunks(originalLines, newLines, lcs, contextLines)
	if len(hunks) > 1 {
		hunks = mergeHunks(originalLines, hunks, contextLines)
	}

	fileDiff := FileDiff{Hunks: hunks}

	if !apply {
		return fileDiff, false, nil
	}

	// Apply: Reconstruct new content from diff to verify, then write automatically
	reconstructed := applyDiff(originalLines, hunks)
	if !linesEqual(reconstructed, newLines) {
		return fileDiff, false, errors.New("diff application mismatch - possible syntax corruption risk")
	}

	// Atomic write: Temp file then rename.
	tempPath := originalPath + ".tmp"
	contentToWrite := strings.Join(reconstructed, "\n")
	if newHasTrailingNewline {
		contentToWrite += "\n"
	}
	if err := os.WriteFile(tempPath, []byte(contentToWrite), 0644); err != nil {
		return fileDiff, false, errors.Wrap(err, "failed to write temp file")
	}
	if err := os.Rename(tempPath, originalPath); err != nil {
		os.Remove(tempPath) // Cleanup.
		return fileDiff, false, errors.Wrap(err, "failed to rename temp file")
	}

	return fileDiff, true, nil
}

// splitLines splits string into lines, preserving trailing newline if any.
func splitLines(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// computeLCS returns LCS length matrix (DP table).
func computeLCS(a, b []string) [][]int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp
}

// buildHunks backtraces the LCS to create unified-style hunks with context.
// Treat changed lines as remove ('-') plus add ('+') pairs; matches are ' '.
func buildHunks(a, b []string, dp [][]int, context int) []DiffHunk {
	// Step 1: Backtrace to build a linear diff of tagged lines.
	type diffLine struct {
		tag  byte // ' ', '-', '+'
		text string
	}
	var reversed []diffLine // we will reverse at the end
	i, j := len(a), len(b)
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			reversed = append(reversed, diffLine{tag: ' ', text: a[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j] == dp[i][j-1]) {
			// Addition from new file.
			reversed = append(reversed, diffLine{tag: '+', text: b[j-1]})
			j--
		} else if i > 0 && (j == 0 || dp[i][j] == dp[i-1][j]) {
			// Removal from original file.
			reversed = append(reversed, diffLine{tag: '-', text: a[i-1]})
			i--
		} else {
			// Fallback (shouldn't occur): treat as removal then addition.
			reversed = append(reversed, diffLine{tag: '-', text: a[i-1]})
			i--
		}
	}
	// Reverse to forward order.
	diff := make([]diffLine, len(reversed))
	for x := range reversed {
		diff[len(reversed)-1-x] = reversed[x]
	}

	// If there are no changes, return no hunks.
	hasChange := false
	for _, dl := range diff {
		if dl.tag == '-' || dl.tag == '+' {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return nil
	}

	// Step 2: Precompute original/new positions for each diff line.
	origPos := make([]int, len(diff)) // 0-based line index in original
	newPos := make([]int, len(diff))  // 0-based line index in new
	o, n := 0, 0
	for idx, dl := range diff {
		origPos[idx] = o
		newPos[idx] = n
		switch dl.tag {
		case ' ':
			o++
			n++
		case '-':
			o++
		case '+':
			n++
		}
	}

	// Step 3: Group into hunks with leading/trailing context.
	var hunks []DiffHunk
	inHunk := false
	hunkStartIdx := 0
	trailing := 0
	for idx, dl := range diff {
		if dl.tag == '-' || dl.tag == '+' {
			if !inHunk {
				inHunk = true
				// Include up to `context` lines before this change.
				if idx-context > 0 {
					hunkStartIdx = idx - context
				} else {
					hunkStartIdx = 0
				}
				trailing = 0
			} else {
				// Reset trailing if more changes occur before we close.
				trailing = 0
			}
		} else if dl.tag == ' ' && inHunk {
			trailing++
			if trailing >= context {
				// Close this hunk including the trailing context lines.
				startOrig := origPos[hunkStartIdx]
				startNew := newPos[hunkStartIdx]
				lines := make([]string, 0, idx-hunkStartIdx+1)
				for k := hunkStartIdx; k <= idx; k++ {
					lines = append(lines, fmt.Sprintf("%c%s", diff[k].tag, diff[k].text))
				}
				hunks = append(hunks, DiffHunk{StartLineOriginal: startOrig, StartLineNew: startNew, Lines: lines})
				inHunk = false
				trailing = 0
			}
		}
	}
	// If still in a hunk at end, close it.
	if inHunk {
		startOrig := origPos[hunkStartIdx]
		startNew := newPos[hunkStartIdx]
		lines := make([]string, 0, len(diff)-hunkStartIdx)
		for k := hunkStartIdx; k < len(diff); k++ {
			lines = append(lines, fmt.Sprintf("%c%s", diff[k].tag, diff[k].text))
		}
		hunks = append(hunks, DiffHunk{StartLineOriginal: startOrig, StartLineNew: startNew, Lines: lines})
	}
	return hunks
}

// mergeHunks merges adjacent hunks if they are separated by fewer than 2*context unchanged lines.
func mergeHunks(a []string, hunks []DiffHunk, context int) []DiffHunk {
	if len(hunks) <= 1 {
		return hunks
	}
	var merged []DiffHunk
	current := hunks[0]
	for i := 1; i < len(hunks); i++ {
		next := hunks[i]
		// Compute end positions in original for current hunk.
		currOrigCount := 0
		for _, ln := range current.Lines {
			if len(ln) == 0 {
				continue
			}
			if ln[0] == ' ' || ln[0] == '-' {
				currOrigCount++
			}
		}
		currEndOrig := current.StartLineOriginal + currOrigCount
		gapOrig := next.StartLineOriginal - currEndOrig
		if gapOrig < 2*context {
			// Merge: include gap lines from original as unchanged and append next.
			gap := make([]string, 0, max(0, gapOrig))
			for k := currEndOrig; k < next.StartLineOriginal && k < len(a); k++ {
				gap = append(gap, " "+a[k])
			}
			current.Lines = append(current.Lines, gap...)
			current.Lines = append(current.Lines, next.Lines...)
			// No change to StartLineNew/Original for current
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)
	return merged
}

// applyDiff reconstructs new lines from original + hunks.
func applyDiff(original []string, hunks []DiffHunk) []string {
	var result []string
	lineIdx := 0
	for _, hunk := range hunks {
		// Add unchanged lines before hunk.
		for lineIdx < hunk.StartLineOriginal {
			result = append(result, original[lineIdx])
			lineIdx++
		}
		for _, line := range hunk.Lines {
			switch line[0] {
			case '+':
				result = append(result, line[1:])
			case ' ':
				result = append(result, line[1:])
				lineIdx++
			case '-':
				lineIdx++ // Skip removed.
			}
		}
	}
	// Add remaining unchanged.
	result = append(result, original[lineIdx:]...)
	return result
}

// FormatDiffForTerminal returns a unified-diff style string suitable for CLI output.
// Example output:
// @@ -startOrig,countOrig +startNew,countNew @@
//
//	context line
//
// -removed line
// +added line
func FormatDiffForTerminal(fd FileDiff) string {
	var sb strings.Builder
	for _, h := range fd.Hunks {
		// Compute counts.
		origCount := 0
		newCount := 0
		for _, ln := range h.Lines {
			if len(ln) == 0 {
				continue
			}
			switch ln[0] {
			case ' ': // context counts in both
				origCount++
				newCount++
			case '-':
				origCount++
			case '+':
				newCount++
			}
		}
		// Unified header uses 1-based line numbers.
		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", h.StartLineOriginal+1, origCount, h.StartLineNew+1, newCount))
		for _, ln := range h.Lines {
			sb.WriteString(ln)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

/*
Example usage:

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
}
*/

// Helper funcs.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func reverseHunks(hunks []DiffHunk) {
	for i, j := 0, len(hunks)-1; i < j; i, j = i+1, j-1 {
		hunks[i], hunks[j] = hunks[j], hunks[i]
	}
}
