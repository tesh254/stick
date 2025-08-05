package vc

// CommitOutput represents the structured output of a commit diff.
type CommitOutput struct {
	Commit struct {
		Hash    string `json:"hash"`
		Author  string `json:"author"`
		Email   string `json:"email"`
		Date    string `json:"date"`
		Message string `json:"message"`
	} `json:"commit"`
	Files []FileChange `json:"files"`
}

// FileChange represents the changes in a single file.
type FileChange struct {
	Path    string       `json:"path"`
	Status  string       `json:"status"`
	Changes []LineChange `json:"changes"`
}

// LineChange represents a single line change in a file.
type LineChange struct {
	Line    int    `json:"line"`
	Type    string `json:"type"` // Type of change: "added" or "deleted"
	Content string `json:"content"`
}
