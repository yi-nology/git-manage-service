package api

import "time"

type AliasEntry struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthorIdentityDTO struct {
	ID             uint         `json:"id"`
	CanonicalName  string       `json:"canonical_name"`
	CanonicalEmail string       `json:"canonical_email"`
	Aliases        []AliasEntry `json:"aliases"`
	IsDefault      bool         `json:"is_default"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type CreateIdentityRequest struct {
	CanonicalName  string       `json:"canonical_name"`
	CanonicalEmail string       `json:"canonical_email"`
	Aliases        []AliasEntry `json:"aliases"`
}

type UpdateIdentityRequest struct {
	CanonicalName  string       `json:"canonical_name"`
	CanonicalEmail string       `json:"canonical_email"`
	Aliases        []AliasEntry `json:"aliases"`
}

type RepoAuthorConfigDTO struct {
	RepoKey    string             `json:"repo_key"`
	IdentityID *uint              `json:"identity_id"`
	Identity   *AuthorIdentityDTO `json:"identity"`
	Source     string             `json:"source"`
}

type MismatchedCommit struct {
	Hash           string `json:"hash"`
	ShortHash      string `json:"short_hash"`
	Message        string `json:"message"`
	AuthorName     string `json:"author_name"`
	AuthorEmail    string `json:"author_email"`
	CommitterName  string `json:"committer_name"`
	CommitterEmail string `json:"committer_email"`
	Date           string `json:"date"`
	MatchType      string `json:"match_type"`
	TargetName     string `json:"target_name"`
	TargetEmail    string `json:"target_email"`
}

type AuthorScanResult struct {
	Commits      []MismatchedCommit `json:"commits"`
	TotalCommits int64              `json:"total_commits"`
	MatchCount   int                `json:"match_count"`
}
