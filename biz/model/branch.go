package model

import "time"

type BranchInfo struct {
	Name        string    `json:"name"`
	IsCurrent   bool      `json:"is_current"`
	Hash        string    `json:"hash"`
	Author      string    `json:"author"`
	AuthorEmail string    `json:"author_email"`
	Date        time.Time `json:"date"`
	Message     string    `json:"message"`
}

type CreateBranchReq struct {
	Name    string `json:"name"` // Removed binding:"required" as hertz might handle it differently or I'll validate manually
	BaseRef string `json:"base_ref"`
}

type UpdateBranchReq struct {
	NewName string `json:"new_name"`
	Desc    string `json:"desc"` // Description is not native to git branch, maybe store in config or ignore for now? Requirement said "Modify branch name and description". Git doesn't strictly support branch description except via `git branch --edit-description` which opens an editor. `git config branch.<name>.description` works.
}
