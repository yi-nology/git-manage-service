package mirror

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	mirrorModel "github.com/yi-nology/git-manage-service/biz/model/mirror"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	mirrorSvc "github.com/yi-nology/git-manage-service/biz/service/mirror"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func getMirrorService() *mirrorSvc.MirrorService {
	return mirrorSvc.GlobalMirrorService
}

func convertToProtoMirror(m po.Mirror) *mirrorModel.Mirror {
	proto := &mirrorModel.Mirror{
		Id:           int64(m.ID),
		RepoId:       int64(m.RepoID),
		MirrorType:   m.MirrorType,
		RemoteUrl:    m.RemoteURL,
		RemoteName:   m.RemoteName,
		BranchFilter: m.BranchFilter,
		SyncInterval: int32(m.SyncInterval),
		CronExpr:     m.CronExpr,
		SyncOnPush:   m.SyncOnPush,
		GitForce:     m.GitForce,
		GitPrune:     m.GitPrune,
		GitTags:      m.GitTags,
		Enabled:      m.Enabled,
		Status:       m.Status,
		LastError:    m.LastError,
		RetryCount:   int32(m.RetryCount),
		WebhookToken: m.WebhookToken,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if m.CredentialID != nil {
		proto.CredentialId = int64(*m.CredentialID)
	}

	if m.LastSyncAt != nil {
		proto.LastSyncAt = m.LastSyncAt.Format("2006-01-02T15:04:05Z")
	}

	if m.NextSyncAt != nil {
		proto.NextSyncAt = m.NextSyncAt.Format("2006-01-02T15:04:05Z")
	}

	if m.Repo.ID != 0 {
		proto.RepoName = m.Repo.Name
		proto.RepoKey = m.Repo.Key
	}

	if m.Credential != nil {
		proto.CredentialName = m.Credential.Name
	}

	return proto
}

func convertToProtoMirrorSyncLog(l po.MirrorSyncLog) *mirrorModel.MirrorSyncLog {
	proto := &mirrorModel.MirrorSyncLog{
		Id:             int64(l.ID),
		MirrorId:       int64(l.MirrorID),
		TriggerType:    l.TriggerType,
		Status:         l.Status,
		DurationMs:     l.DurationMs,
		BranchesSynced: int32(l.BranchesSynced),
		CommitsPushed:  int32(l.CommitsPushed),
		ErrorMessage:   l.ErrorMessage,
		DetailLog:      l.DetailLog,
		CreatedAt:      l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if l.StartedAt != nil {
		proto.StartedAt = l.StartedAt.Format("2006-01-02T15:04:05Z")
	}

	if l.FinishedAt != nil {
		proto.FinishedAt = l.FinishedAt.Format("2006-01-02T15:04:05Z")
	}

	return proto
}

func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02T15:04:05Z", s)
	if err != nil {
		return nil
	}
	return &t
}

func parseUintPtr(v int64) *uint {
	if v == 0 {
		return nil
	}
	u := uint(v)
	return &u
}

func ListMirrors(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	repoIDStr := c.Query("repo_id")
	mirrorType := c.Query("mirror_type")

	var mirrors []po.Mirror
	var err error

	if repoIDStr != "" {
		repoID, parseErr := strconv.ParseUint(repoIDStr, 10, 64)
		if parseErr != nil {
			response.BadRequest(c, "invalid repo_id")
			return
		}
		mirrors, err = svc.ListMirrorsByRepo(uint(repoID))
	} else if mirrorType != "" {
		mirrors, err = svc.ListMirrors()
		if err == nil {
			filtered := make([]po.Mirror, 0)
			for _, m := range mirrors {
				if m.MirrorType == mirrorType {
					filtered = append(filtered, m)
				}
			}
			mirrors = filtered
		}
	} else {
		mirrors, err = svc.ListMirrors()
	}

	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	dtos := make([]*mirrorModel.Mirror, 0, len(mirrors))
	for _, m := range mirrors {
		dtos = append(dtos, convertToProtoMirror(m))
	}
	response.Success(c, dtos)
}

func GetMirror(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mirror, err := svc.GetMirror(id)
	if err != nil {
		response.NotFound(c, "mirror not found")
		return
	}
	response.Success(c, convertToProtoMirror(*mirror))
}

func CreateMirror(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	var req api.CreateMirrorReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.MirrorType != po.MirrorTypePull && req.MirrorType != po.MirrorTypePush {
		response.BadRequest(c, "mirror_type must be 'pull' or 'push'")
		return
	}

	mirror := &po.Mirror{
		RepoID:       req.RepoID,
		MirrorType:   req.MirrorType,
		RemoteURL:    req.RemoteURL,
		RemoteName:   req.RemoteName,
		CredentialID: req.CredentialID,
		BranchFilter: req.BranchFilter,
		SyncInterval: req.SyncInterval,
		CronExpr:     req.CronExpr,
		SyncOnPush:   req.SyncOnPush,
		GitForce:     req.GitForce,
		GitPrune:     req.GitPrune,
		GitTags:      req.GitTags,
		Enabled:      req.Enabled,
	}

	if err := svc.CreateMirror(mirror); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	created, _ := svc.GetMirror(mirror.ID)
	if created != nil {
		response.Success(c, convertToProtoMirror(*created))
	} else {
		response.Success(c, convertToProtoMirror(*mirror))
	}

	if mirrorSvc.GlobalScheduler != nil && mirror.CronExpr != "" {
		mirrorSvc.GlobalScheduler.AddCronMirror(mirror)
	}
}

func UpdateMirror(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req api.UpdateMirrorReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mirror, err := svc.GetMirror(id)
	if err != nil {
		response.NotFound(c, "mirror not found")
		return
	}

	if req.RemoteURL != "" {
		mirror.RemoteURL = req.RemoteURL
	}
	mirror.RemoteName = req.RemoteName
	mirror.CredentialID = req.CredentialID
	mirror.BranchFilter = req.BranchFilter
	if req.SyncInterval > 0 {
		mirror.SyncInterval = req.SyncInterval
	}
	mirror.CronExpr = req.CronExpr
	mirror.SyncOnPush = req.SyncOnPush
	mirror.GitForce = req.GitForce
	mirror.GitPrune = req.GitPrune
	mirror.GitTags = req.GitTags
	mirror.Enabled = req.Enabled

	if err := svc.UpdateMirror(mirror); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	updated, _ := svc.GetMirror(id)
	if updated != nil {
		response.Success(c, convertToProtoMirror(*updated))
	} else {
		response.Success(c, convertToProtoMirror(*mirror))
	}

	if mirrorSvc.GlobalScheduler != nil {
		if mirror.CronExpr != "" && mirror.Enabled {
			mirrorSvc.GlobalScheduler.AddCronMirror(mirror)
		} else {
			mirrorSvc.GlobalScheduler.RemoveCronMirror(mirror.ID)
		}
	}
}

func DeleteMirror(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := svc.DeleteMirror(id); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	if mirrorSvc.GlobalScheduler != nil {
		mirrorSvc.GlobalScheduler.RemoveCronMirror(id)
	}

	response.Success(c, nil)
}

func TriggerSync(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	triggerType := c.DefaultPostForm("trigger_type", po.TriggerTypeManual)

	if err := svc.TriggerSync(id, triggerType); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Accepted(c, "sync triggered", nil)
}

func BatchTriggerSync(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	var req struct {
		MirrorIDs   []uint `json:"mirror_ids"`
		TriggerType string `json:"trigger_type"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.TriggerType == "" {
		req.TriggerType = po.TriggerTypeManual
	}

	if err := svc.BatchTriggerSync(req.MirrorIDs, req.TriggerType); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Accepted(c, "batch sync triggered", nil)
}

func PreviewSync(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	preview, err := svc.PreviewSync(ctx, id)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"preview": preview})
}

func ListSyncLogs(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, e := strconv.Atoi(l); e == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := svc.ListSyncLogs(id, limit)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	dtos := make([]*mirrorModel.MirrorSyncLog, 0, len(logs))
	for _, l := range logs {
		dtos = append(dtos, convertToProtoMirrorSyncLog(l))
	}
	response.Success(c, dtos)
}

func GetSyncLog(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	logID, err := parseIDParam(c, "log_id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	log, err := svc.GetSyncLog(logID)
	if err != nil {
		response.NotFound(c, "sync log not found")
		return
	}
	response.Success(c, convertToProtoMirrorSyncLog(*log))
}

func DeleteSyncLog(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	logID, err := parseIDParam(c, "log_id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := svc.DeleteSyncLog(logID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func PauseMirror(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := svc.PauseMirror(id); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	if mirrorSvc.GlobalScheduler != nil {
		mirrorSvc.GlobalScheduler.RemoveCronMirror(id)
	}

	mirror, _ := svc.GetMirror(id)
	if mirror != nil {
		response.Success(c, convertToProtoMirror(*mirror))
	} else {
		response.Success(c, nil)
	}
}

func ResumeMirror(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := svc.ResumeMirror(id); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	mirror, _ := svc.GetMirror(id)
	if mirror != nil && mirror.CronExpr != "" && mirrorSvc.GlobalScheduler != nil {
		mirrorSvc.GlobalScheduler.AddCronMirror(mirror)
	}

	if mirror != nil {
		response.Success(c, convertToProtoMirror(*mirror))
	} else {
		response.Success(c, nil)
	}
}

func AnalyzeRemote(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	var req struct {
		RemoteURL    string `json:"remote_url"`
		CredentialID int64  `json:"credential_id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := svc.AnalyzeRemote(ctx, req.RemoteURL)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func ValidateCredential(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	var req struct {
		CredentialID uint   `json:"credential_id"`
		RemoteURL    string `json:"remote_url"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"valid":   true,
		"message": "credential validation not yet implemented",
	})
}

func HandleWebhook(ctx context.Context, c *app.RequestContext) {
	svc := getMirrorService()
	if svc == nil {
		response.InternalServerError(c, "mirror service not initialized")
		return
	}

	token := c.Param("token")
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	if err := svc.HandleWebhook(token); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func parseIDParam(c *app.RequestContext, param string) (uint, error) {
	str := c.Param(param)
	if str == "" {
		return 0, strconv.ErrSyntax
	}
	id, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
