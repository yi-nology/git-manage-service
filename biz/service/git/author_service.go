package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type AuthorService struct{}

func NewAuthorService() *AuthorService {
	return &AuthorService{}
}

func (s *AuthorService) ListIdentities() ([]api.AuthorIdentityDTO, error) {
	identities, err := db.NewAuthorIdentityDAO().ListAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]api.AuthorIdentityDTO, 0, len(identities))
	for _, id := range identities {
		dtos = append(dtos, identityToDTO(&id))
	}
	return dtos, nil
}

func (s *AuthorService) CreateIdentity(req api.CreateIdentityRequest) (*api.AuthorIdentityDTO, error) {
	aliasesJSON, _ := json.Marshal(req.Aliases)
	identity := &po.AuthorIdentity{
		CanonicalName:  req.CanonicalName,
		CanonicalEmail: req.CanonicalEmail,
		AliasesJSON:    string(aliasesJSON),
	}
	dao := db.NewAuthorIdentityDAO()
	if err := dao.Create(identity); err != nil {
		return nil, err
	}
	dto := identityToDTO(identity)
	return &dto, nil
}

func (s *AuthorService) UpdateIdentity(id uint, req api.UpdateIdentityRequest) (*api.AuthorIdentityDTO, error) {
	dao := db.NewAuthorIdentityDAO()
	identity, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}
	identity.CanonicalName = req.CanonicalName
	identity.CanonicalEmail = req.CanonicalEmail
	aliasesJSON, _ := json.Marshal(req.Aliases)
	identity.AliasesJSON = string(aliasesJSON)
	if err := dao.Update(identity); err != nil {
		return nil, err
	}
	dto := identityToDTO(identity)
	return &dto, nil
}

func (s *AuthorService) DeleteIdentity(id uint) error {
	return db.NewAuthorIdentityDAO().Delete(id)
}

func (s *AuthorService) ActivateIdentity(id uint) error {
	dao := db.NewAuthorIdentityDAO()
	identity, err := dao.FindByID(id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}
	if err := dao.SetDefault(id); err != nil {
		return err
	}
	gitSvc := NewGitService()
	return gitSvc.SetGlobalGitUser(identity.CanonicalName, identity.CanonicalEmail)
}

func (s *AuthorService) GetRepoAuthorConfig(repoKey string) (*api.RepoAuthorConfigDTO, error) {
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	cfg := &api.RepoAuthorConfigDTO{RepoKey: repoKey}
	if repo.AuthorIdentityID != nil {
		identity, err := db.NewAuthorIdentityDAO().FindByID(*repo.AuthorIdentityID)
		if err == nil {
			dto := identityToDTO(identity)
			cfg.IdentityID = repo.AuthorIdentityID
			cfg.Identity = &dto
			cfg.Source = "repo"
			return cfg, nil
		}
	}
	defaultIdentity, err := db.NewAuthorIdentityDAO().GetDefault()
	if err == nil {
		dto := identityToDTO(defaultIdentity)
		cfg.Identity = &dto
		cfg.IdentityID = &defaultIdentity.ID
		cfg.Source = "global"
	}
	return cfg, nil
}

func (s *AuthorService) SetRepoAuthorConfig(repoKey string, identityID *uint, clear bool) error {
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return fmt.Errorf("repo not found: %w", err)
	}
	if clear {
		repo.AuthorIdentityID = nil
	} else {
		repo.AuthorIdentityID = identityID
	}
	return db.NewRepoDAO().Save(repo)
}

func (s *AuthorService) ScanAuthor(repoPath string) (*api.AuthorScanResult, error) {
	allIdentities, err := db.NewAuthorIdentityDAO().ListAll()
	if err != nil {
		return nil, err
	}
	type aliasInfo struct {
		name           string
		email          string
		canonicalName  string
		canonicalEmail string
	}
	var aliases []aliasInfo
	for _, id := range allIdentities {
		var aliasEntries []api.AliasEntry
		json.Unmarshal([]byte(id.AliasesJSON), &aliasEntries)
		for _, a := range aliasEntries {
			aliases = append(aliases, aliasInfo{
				name:           a.Name,
				email:          a.Email,
				canonicalName:  id.CanonicalName,
				canonicalEmail: id.CanonicalEmail,
			})
		}
	}
	cmd := exec.Command("git", "log", "--all", "--format=%H|%h|%an|%ae|%cn|%ce|%ai|%s")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}
	var mismatched []api.MismatchedCommit
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.SplitN(line, "|", 8)
		if len(fields) < 8 {
			continue
		}
		hash := fields[0]
		shortHash := fields[1]
		authorName := fields[2]
		authorEmail := fields[3]
		committerName := fields[4]
		committerEmail := fields[5]
		dateStr := fields[6]
		message := fields[7]

		var match *api.MismatchedCommit
		for _, a := range aliases {
			if authorName == a.name && authorEmail == a.email {
				match = &api.MismatchedCommit{
					Hash: hash, ShortHash: shortHash, Message: message,
					AuthorName: authorName, AuthorEmail: authorEmail,
					CommitterName: committerName, CommitterEmail: committerEmail,
					Date: dateStr, MatchType: "exact",
					TargetName: a.canonicalName, TargetEmail: a.canonicalEmail,
				}
				break
			}
			if authorEmail == a.email {
				match = &api.MismatchedCommit{
					Hash: hash, ShortHash: shortHash, Message: message,
					AuthorName: authorName, AuthorEmail: authorEmail,
					CommitterName: committerName, CommitterEmail: committerEmail,
					Date: dateStr, MatchType: "email_only",
					TargetName: a.canonicalName, TargetEmail: a.canonicalEmail,
				}
			}
		}
		if match != nil {
			mismatched = append(mismatched, *match)
		}
	}
	return &api.AuthorScanResult{
		Commits:      mismatched,
		TotalCommits: int64(len(lines)),
		MatchCount:   len(mismatched),
	}, nil
}

func (s *AuthorService) FixAuthorAll(repoPath string, pushRemote string, taskID string) error {
	return s.fixAuthor(repoPath, nil, pushRemote, taskID)
}

func (s *AuthorService) FixAuthor(repoPath string, commitHashes []string, pushRemote string, taskID string) error {
	return s.fixAuthor(repoPath, commitHashes, pushRemote, taskID)
}

func (s *AuthorService) fixAuthor(repoPath string, commitHashes []string, pushRemote string, taskID string) error {
	tm := GlobalTaskManager

	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
	}

	allIdentities, err := db.NewAuthorIdentityDAO().ListAll()
	if err != nil {
		return err
	}

	type aliasMapping struct {
		oldName  string
		oldEmail string
		newName  string
		newEmail string
	}
	var mappings []aliasMapping
	for _, id := range allIdentities {
		var aliasEntries []api.AliasEntry
		json.Unmarshal([]byte(id.AliasesJSON), &aliasEntries)
		for _, a := range aliasEntries {
			mappings = append(mappings, aliasMapping{
				oldName:  a.Name,
				oldEmail: a.Email,
				newName:  id.CanonicalName,
				newEmail: id.CanonicalEmail,
			})
		}
	}

	if len(mappings) == 0 {
		appendLog("没有别名映射，无需修复")
		return nil
	}

	appendLog("构建 env-filter 脚本...")
	var filterParts []string
	for _, m := range mappings {
		filterParts = append(filterParts, fmt.Sprintf(
			`if [ "$GIT_AUTHOR_EMAIL" = "%s" ] && [ "$GIT_AUTHOR_NAME" = "%s" ]; then export GIT_AUTHOR_NAME="%s"; export GIT_AUTHOR_EMAIL="%s"; fi`,
			m.oldEmail, m.oldName, m.newName, m.newEmail,
		))
		filterParts = append(filterParts, fmt.Sprintf(
			`if [ "$GIT_AUTHOR_EMAIL" = "%s" ]; then export GIT_AUTHOR_NAME="%s"; export GIT_AUTHOR_EMAIL="%s"; fi`,
			m.oldEmail, m.newName, m.newEmail,
		))
		filterParts = append(filterParts, fmt.Sprintf(
			`if [ "$GIT_COMMITTER_EMAIL" = "%s" ] && [ "$GIT_COMMITTER_NAME" = "%s" ]; then export GIT_COMMITTER_NAME="%s"; export GIT_COMMITTER_EMAIL="%s"; fi`,
			m.oldEmail, m.oldName, m.newName, m.newEmail,
		))
		filterParts = append(filterParts, fmt.Sprintf(
			`if [ "$GIT_COMMITTER_EMAIL" = "%s" ]; then export GIT_COMMITTER_NAME="%s"; export GIT_COMMITTER_EMAIL="%s"; fi`,
			m.oldEmail, m.newName, m.newEmail,
		))
	}

	envFilter := strings.Join(filterParts, "\n")

	appendLog("清理 backup refs...")
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/original/")
	cmd.Dir = repoPath
	refsOutput, _ := cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			exec.Command("git", "update-ref", "-d", ref).Run()
		}
	}

	appendLog("执行 filter-branch 重写作者信息...")
	args := []string{"filter-branch", "--force", "--env-filter", envFilter, "--tag-name-filter", "cat", "--", "--all"}
	cmd = exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("filter-branch failed: %w, output: %s", err, string(output))
	}
	appendLog("filter-branch 完成")

	appendLog("清理 refs/original/ ...")
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/original/")
	cmd.Dir = repoPath
	refsOutput, _ = cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			exec.Command("git", "update-ref", "-d", ref).Run()
		}
	}

	appendLog("清理 reflog...")
	exec.Command("git", "reflog", "expire", "--expire=now", "--all").Run()

	appendLog("执行 gc...")
	cmd = exec.Command("git", "gc", "--prune=now")
	cmd.Dir = repoPath
	cmd.CombinedOutput()

	if pushRemote != "" {
		appendLog("推送 " + pushRemote + " (force)...")
		cmd = exec.Command("git", "push", "--force", pushRemote, "--all")
		cmd.Dir = repoPath
		if out, err := cmd.CombinedOutput(); err != nil {
			appendLog("push all 警告: " + string(out))
		} else {
			appendLog("push --all 完成")
		}
		cmd = exec.Command("git", "push", "--force", pushRemote, "--tags")
		cmd.Dir = repoPath
		if out, err := cmd.CombinedOutput(); err != nil {
			appendLog("push tags 警告: " + string(out))
		} else {
			appendLog("push --tags 完成")
		}
	}

	appendLog("作者修复完成！")
	return nil
}

func (s *AuthorService) GetEffectiveAuthor(repoID uint) (name string, email string) {
	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return "", ""
	}
	if repo.AuthorIdentityID != nil {
		identity, err := db.NewAuthorIdentityDAO().FindByID(*repo.AuthorIdentityID)
		if err == nil {
			return identity.CanonicalName, identity.CanonicalEmail
		}
	}
	identity, err := db.NewAuthorIdentityDAO().GetDefault()
	if err == nil {
		return identity.CanonicalName, identity.CanonicalEmail
	}
	return "", ""
}

func StartAuthorFixTask(repoID uint, repoPath string, commitHashes []string, pushRemote string) (string, error) {
	taskID := uuid.New().String()
	GlobalTaskManager.AddTask(taskID)

	record, err := CreateMaintenanceRecord(repoID, "author_fix", repoPath)
	if err != nil {
		return "", err
	}
	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	dao.Update(record)

	svc := NewAuthorService()
	go func() {
		var fixErr error
		if len(commitHashes) == 0 {
			fixErr = svc.FixAuthorAll(repoPath, pushRemote, taskID)
		} else {
			fixErr = svc.FixAuthor(repoPath, commitHashes, pushRemote, taskID)
		}
		now := time.Now()
		if fixErr != nil {
			GlobalTaskManager.UpdateStatus(taskID, "failed", fixErr.Error())
			rec, _ := dao.FindByTaskID(taskID)
			if rec != nil {
				rec.Status = "failed"
				rec.ErrorMessage = fixErr.Error()
				rec.FinishedAt = &now
				dao.Update(rec)
			}
		} else {
			GlobalTaskManager.UpdateStatus(taskID, "success", "")
			rec, _ := dao.FindByTaskID(taskID)
			if rec != nil {
				rec.Status = "success"
				rec.FinishedAt = &now
				dao.Update(rec)
			}
		}
	}()

	return taskID, nil
}

func identityToDTO(id *po.AuthorIdentity) api.AuthorIdentityDTO {
	var aliases []api.AliasEntry
	if id.AliasesJSON != "" {
		json.Unmarshal([]byte(id.AliasesJSON), &aliases)
	}
	if aliases == nil {
		aliases = []api.AliasEntry{}
	}
	return api.AuthorIdentityDTO{
		ID:             id.ID,
		CanonicalName:  id.CanonicalName,
		CanonicalEmail: id.CanonicalEmail,
		Aliases:        aliases,
		IsDefault:      id.IsDefault,
		CreatedAt:      id.CreatedAt,
		UpdatedAt:      id.UpdatedAt,
	}
}
