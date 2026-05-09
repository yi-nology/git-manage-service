package api

type WorkspaceDiffFile struct {
	File      string `json:"file"`
	Diff      string `json:"diff"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	IsBinary  bool   `json:"isBinary"`
}

type WorkspaceDiff struct {
	Files          []WorkspaceDiffFile `json:"files"`
	TotalAdditions int                 `json:"totalAdditions"`
	TotalDeletions int                 `json:"totalDeletions"`
}

type CommitResult struct {
	CommitHash string `json:"commitHash"`
	Pushed     bool   `json:"pushed"`
}

type AIResolvedFile struct {
	FilePath        string  `json:"filePath"`
	ResolvedContent string  `json:"resolvedContent"`
	Explanation     string  `json:"explanation"`
	Confidence      float64 `json:"confidence"`
}
