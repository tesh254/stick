package fileupdater

// DiffHunk represents a group of changes with context.
type DiffHunk struct {
	StartLineOriginal int
	StartLineNew      int
	Lines             []string // Lines with prefixes: '+' added, '-' removed, ' ' unchanged.
}

// FileDiff represents the full diff
type FileDiff struct {
	Hunks []DiffHunk
}
