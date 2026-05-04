package api

import "time"

type AliasEntry struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthorIdentityDTO struct {
	ID             uint         `json:"id"`
	CanonicalName  string       `json:"canonicalName"`
	CanonicalEmail string       `json:"canonicalEmail"`
	Aliases        []AliasEntry `json:"aliases"`
	IsDefault      bool         `json:"isDefault"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type CreateIdentityRequest struct {
	CanonicalName  string       `json:"canonicalName"`
	CanonicalEmail string       `json:"canonicalEmail"`
	Aliases        []AliasEntry `json:"aliases"`
}

type UpdateIdentityRequest struct {
	CanonicalName  string       `json:"canonicalName"`
	CanonicalEmail string       `json:"canonicalEmail"`
	Aliases        []AliasEntry `json:"aliases"`
}

type RepoAuthorConfigDTO struct {
	RepoKey    string             `json:"repoKey"`
	IdentityID *uint              `json:"identityId"`
	Identity   *AuthorIdentityDTO `json:"identity"`
	Source     string             `json:"source"`
}

type MismatchedCommit struct {
	Hash           string `json:"hash"`
	ShortHash      string `json:"shortHash"`
	Message        string `json:"message"`
	AuthorName     string `json:"authorName"`
	AuthorEmail    string `json:"authorEmail"`
	CommitterName  string `json:"committerName"`
	CommitterEmail string `json:"committerEmail"`
	Date           string `json:"date"`
	MatchType      string `json:"matchType"`
	TargetName     string `json:"targetName"`
	TargetEmail    string `json:"targetEmail"`
}

type AuthorScanResult struct {
	Commits      []MismatchedCommit `json:"commits"`
	TotalCommits int64              `json:"totalCommits"`
	MatchCount   int                `json:"matchCount"`
}

type AuthorFixRequest struct {
	CommitHashes []string `json:"commitHashes"`
	PushRemote   string   `json:"pushRemote"`
}

type AuthorFixAllRequest struct {
	PushRemote string `json:"pushRemote"`
}
