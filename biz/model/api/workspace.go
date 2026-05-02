package api

type WorkspaceFileStatus struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	OldPath   string `json:"oldPath,omitempty"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	IsBinary  bool   `json:"isBinary"`
}

type WorkspaceStatus struct {
	Branch     string               `json:"branch"`
	Staged     []WorkspaceFileStatus `json:"staged"`
	Unstaged   []WorkspaceFileStatus `json:"unstaged"`
	Untracked  []WorkspaceFileStatus `json:"untracked"`
	Conflicted []WorkspaceFileStatus `json:"conflicted"`
	IsClean    bool                 `json:"isClean"`
	IsMerging  bool                 `json:"isMerging"`
	IsRebasing bool                 `json:"isRebasing"`
	Ahead      int                  `json:"ahead"`
	Behind     int                  `json:"behind"`
}

type WorkspaceDiffFile struct {
	File       string `json:"file"`
	Diff       string `json:"diff"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	IsBinary   bool   `json:"isBinary"`
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

type PullResult struct {
	Status       string   `json:"status"`
	Conflicts    []string `json:"conflicts"`
	FetchLog     string   `json:"fetchLog"`
	BehindPulled bool     `json:"behindPulled"`
}

type ConflictDetail struct {
	Path           string `json:"path"`
	OursContent    string `json:"oursContent"`
	TheirsContent  string `json:"theirsContent"`
	BaseContent    string `json:"baseContent"`
	ConflictMarker string `json:"conflictMarker"`
}

type AIResolvedFile struct {
	FilePath        string  `json:"filePath"`
	ResolvedContent string  `json:"resolvedContent"`
	Explanation     string  `json:"explanation"`
	Confidence      float64 `json:"confidence"`
}
