package api

type WorkspaceDiffFile struct {
	File      string `json:"file"`
	Diff      string `json:"diff"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	IsBinary  bool   `json:"is_binary"`
}

type WorkspaceDiff struct {
	Files          []WorkspaceDiffFile `json:"files"`
	TotalAdditions int                 `json:"total_additions"`
	TotalDeletions int                 `json:"total_deletions"`
}

type CommitResult struct {
	CommitHash string `json:"commit_hash"`
	Pushed     bool   `json:"pushed"`
}

type AIResolvedFile struct {
	FilePath        string  `json:"file_path"`
	ResolvedContent string  `json:"resolved_content"`
	Explanation     string  `json:"explanation"`
	Confidence      float64 `json:"confidence"`
}
