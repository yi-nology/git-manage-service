package api

type CreateMirrorReq struct {
	RepoID       uint   `json:"repoId" binding:"required"`
	MirrorType   string `json:"mirrorType" binding:"required"`
	RemoteURL    string `json:"remoteUrl" binding:"required"`
	RemoteName   string `json:"remoteName"`
	CredentialID *uint  `json:"credentialId"`
	BranchFilter string `json:"branchFilter"`
	SyncInterval int    `json:"syncInterval"`
	CronExpr     string `json:"cronExpr"`
	SyncOnPush   bool   `json:"syncOnPush"`
	GitForce     bool   `json:"gitForce"`
	GitPrune     bool   `json:"gitPrune"`
	GitTags      bool   `json:"gitTags"`
	Enabled      bool   `json:"enabled"`
}

type UpdateMirrorReq struct {
	RemoteURL    string `json:"remoteUrl"`
	RemoteName   string `json:"remoteName"`
	CredentialID *uint  `json:"credentialId"`
	BranchFilter string `json:"branchFilter"`
	SyncInterval int    `json:"syncInterval"`
	CronExpr     string `json:"cronExpr"`
	SyncOnPush   bool   `json:"syncOnPush"`
	GitForce     bool   `json:"gitForce"`
	GitPrune     bool   `json:"gitPrune"`
	GitTags      bool   `json:"gitTags"`
	Enabled      bool   `json:"enabled"`
}
