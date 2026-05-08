package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MirrorDTO struct {
	ID             uint       `json:"id"`
	RepoID         uint       `json:"repoId"`
	RepoName       string     `json:"repoName"`
	RepoKey        string     `json:"repoKey"`
	MirrorType     string     `json:"mirrorType"`
	RemoteURL      string     `json:"remoteUrl"`
	RemoteName     string     `json:"remoteName"`
	CredentialID   *uint      `json:"credentialId"`
	CredentialName string     `json:"credentialName,omitempty"`
	BranchFilter   string     `json:"branchFilter"`
	SyncInterval   int        `json:"syncInterval"`
	CronExpr       string     `json:"cronExpr"`
	SyncOnPush     bool       `json:"syncOnPush"`
	GitForce       bool       `json:"gitForce"`
	GitPrune       bool       `json:"gitPrune"`
	GitTags        bool       `json:"gitTags"`
	Enabled        bool       `json:"enabled"`
	Status         string     `json:"status"`
	LastSyncAt     *time.Time `json:"lastSyncAt"`
	LastError      string     `json:"lastError"`
	RetryCount     int        `json:"retryCount"`
	NextSyncAt     *time.Time `json:"nextSyncAt"`
	WebhookToken   string     `json:"webhookToken"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type MirrorSyncLogDTO struct {
	ID             uint       `json:"id"`
	MirrorID       uint       `json:"mirrorId"`
	TriggerType    string     `json:"triggerType"`
	Status         string     `json:"status"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	DurationMs     int64      `json:"durationMs"`
	BranchesSynced int        `json:"branchesSynced"`
	CommitsPushed  int        `json:"commitsPushed"`
	ErrorMessage   string     `json:"errorMessage"`
	DetailLog      string     `json:"detailLog,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

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

func NewMirrorDTO(m po.Mirror) MirrorDTO {
	dto := MirrorDTO{
		ID:           m.ID,
		RepoID:       m.RepoID,
		MirrorType:   m.MirrorType,
		RemoteURL:    m.RemoteURL,
		RemoteName:   m.RemoteName,
		CredentialID: m.CredentialID,
		BranchFilter: m.BranchFilter,
		SyncInterval: m.SyncInterval,
		CronExpr:     m.CronExpr,
		SyncOnPush:   m.SyncOnPush,
		GitForce:     m.GitForce,
		GitPrune:     m.GitPrune,
		GitTags:      m.GitTags,
		Enabled:      m.Enabled,
		Status:       m.Status,
		LastSyncAt:   m.LastSyncAt,
		LastError:    m.LastError,
		RetryCount:   m.RetryCount,
		NextSyncAt:   m.NextSyncAt,
		WebhookToken: m.WebhookToken,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}

	if m.Repo.ID != 0 {
		dto.RepoName = m.Repo.Name
		dto.RepoKey = m.Repo.Key
	}

	if m.Credential != nil {
		dto.CredentialName = m.Credential.Name
	}

	return dto
}

func NewMirrorSyncLogDTO(l po.MirrorSyncLog) MirrorSyncLogDTO {
	return MirrorSyncLogDTO{
		ID:             l.ID,
		MirrorID:       l.MirrorID,
		TriggerType:    l.TriggerType,
		Status:         l.Status,
		StartedAt:      l.StartedAt,
		FinishedAt:     l.FinishedAt,
		DurationMs:     l.DurationMs,
		BranchesSynced: l.BranchesSynced,
		CommitsPushed:  l.CommitsPushed,
		ErrorMessage:   l.ErrorMessage,
		DetailLog:      l.DetailLog,
		CreatedAt:      l.CreatedAt,
	}
}
