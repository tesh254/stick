package metadata

type Metadata struct {
	VirtualBranches map[string]VirtualBranch `json:"virtualBranches"`
}

type VirtualBranch struct {
	GitBranch string   `json:"gitBranch"`
	Files     []string `json:"files"`
}
