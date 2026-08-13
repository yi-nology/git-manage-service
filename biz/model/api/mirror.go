package api

type CreateMirrorReq struct {
	RepoID       uint   `json:"repo_id" binding:"required"`
	MirrorType   string `json:"mirror_type" binding:"required"`
	RemoteURL    string `json:"remote_url" binding:"required"`
	RemoteName   string `json:"remote_name"`
	CredentialID *uint  `json:"credential_id"`
	BranchFilter string `json:"branch_filter"`
	SyncInterval int    `json:"sync_interval"`
	CronExpr     string `json:"cron_expr"`
	SyncOnPush   bool   `json:"sync_on_push"`
	GitForce     bool   `json:"git_force"`
	GitPrune     bool   `json:"git_prune"`
	GitTags      bool   `json:"git_tags"`
	Enabled      bool   `json:"enabled"`
}

type UpdateMirrorReq struct {
	RemoteURL    string `json:"remote_url"`
	RemoteName   string `json:"remote_name"`
	CredentialID *uint  `json:"credential_id"`
	BranchFilter string `json:"branch_filter"`
	SyncInterval int    `json:"sync_interval"`
	CronExpr     string `json:"cron_expr"`
	SyncOnPush   bool   `json:"sync_on_push"`
	GitForce     bool   `json:"git_force"`
	GitPrune     bool   `json:"git_prune"`
	GitTags      bool   `json:"git_tags"`
	Enabled      bool   `json:"enabled"`
}
