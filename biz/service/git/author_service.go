package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

var authorFixMu sync.Mutex

type AuthorService struct{}

func NewAuthorService() *AuthorService {
	return &AuthorService{}
}

func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func runGit(repoPath string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd
}

func runGitWithEnv(repoPath string, env []string, args ...string) *exec.Cmd {
	cmd := runGit(repoPath, args...)
	cmd.Env = env
	return cmd
}

func validateIdentity(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("名称不能为空")
	}
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("邮箱不能为空")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确: %s", email)
	}
	return nil
}

func (s *AuthorService) ListIdentities() ([]api.AuthorIdentityDTO, error) {
	identities, err := db.NewAuthorIdentityDAO().ListAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]api.AuthorIdentityDTO, 0, len(identities))
	for _, id := range identities {
		dtos = append(dtos, IdentityToDTO(&id))
	}
	return dtos, nil
}

func (s *AuthorService) CreateIdentity(req api.CreateIdentityRequest) (*api.AuthorIdentityDTO, error) {
	if err := validateIdentity(req.CanonicalName, req.CanonicalEmail); err != nil {
		return nil, fmt.Errorf("参数校验失败: %w", err)
	}
	for _, a := range req.Aliases {
		if err := validateIdentity(a.Name, a.Email); err != nil {
			return nil, fmt.Errorf("别名校验失败: %w", err)
		}
	}
	aliasesJSON, err := json.Marshal(req.Aliases)
	if err != nil {
		return nil, fmt.Errorf("序列化别名失败: %w", err)
	}
	identity := &po.AuthorIdentity{
		CanonicalName:  req.CanonicalName,
		CanonicalEmail: req.CanonicalEmail,
		AliasesJSON:    string(aliasesJSON),
	}
	dao := db.NewAuthorIdentityDAO()
	if err := dao.Create(identity); err != nil {
		return nil, err
	}
	dto := IdentityToDTO(identity)
	return &dto, nil
}

func (s *AuthorService) UpdateIdentity(id uint, req api.UpdateIdentityRequest) (*api.AuthorIdentityDTO, error) {
	if err := validateIdentity(req.CanonicalName, req.CanonicalEmail); err != nil {
		return nil, fmt.Errorf("参数校验失败: %w", err)
	}
	for _, a := range req.Aliases {
		if err := validateIdentity(a.Name, a.Email); err != nil {
			return nil, fmt.Errorf("别名校验失败: %w", err)
		}
	}
	dao := db.NewAuthorIdentityDAO()
	identity, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}
	identity.CanonicalName = req.CanonicalName
	identity.CanonicalEmail = req.CanonicalEmail
	aliasesJSON, err := json.Marshal(req.Aliases)
	if err != nil {
		return nil, fmt.Errorf("序列化别名失败: %w", err)
	}
	identity.AliasesJSON = string(aliasesJSON)
	if err := dao.Update(identity); err != nil {
		return nil, err
	}
	dto := IdentityToDTO(identity)
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
	if err := gitSvc.SetGlobalGitUser(identity.CanonicalName, identity.CanonicalEmail); err != nil {
		_ = dao.SetDefault(0)
		return fmt.Errorf("更新 ~/.gitconfig 失败: %w", err)
	}
	return nil
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
			dto := IdentityToDTO(identity)
			cfg.IdentityID = repo.AuthorIdentityID
			cfg.Identity = &dto
			cfg.Source = "repo"
			return cfg, nil
		}
	}
	defaultIdentity, err := db.NewAuthorIdentityDAO().GetDefault()
	if err == nil {
		dto := IdentityToDTO(defaultIdentity)
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

type aliasInfo struct {
	name           string
	email          string
	canonicalName  string
	canonicalEmail string
}

func loadAllAliases() ([]aliasInfo, error) {
	allIdentities, err := db.NewAuthorIdentityDAO().ListAll()
	if err != nil {
		return nil, err
	}
	var aliases []aliasInfo
	for _, id := range allIdentities {
		var aliasEntries []api.AliasEntry
		if err := json.Unmarshal([]byte(id.AliasesJSON), &aliasEntries); err != nil {
			continue
		}
		for _, a := range aliasEntries {
			aliases = append(aliases, aliasInfo{
				name:           a.Name,
				email:          a.Email,
				canonicalName:  id.CanonicalName,
				canonicalEmail: id.CanonicalEmail,
			})
		}
	}
	return aliases, nil
}

func buildAliasIndex(aliases []aliasInfo) map[string][]aliasInfo {
	idx := make(map[string][]aliasInfo)
	for _, a := range aliases {
		idx[a.email] = append(idx[a.email], a)
	}
	return idx
}

func (s *AuthorService) ScanAuthor(repoPath string) (*api.AuthorScanResult, error) {
	aliases, err := loadAllAliases()
	if err != nil {
		return nil, err
	}
	emailIndex := buildAliasIndex(aliases)

	cmd := exec.Command("git", "log", "--all",
		"--format=%H%x00%h%x00%an%x00%ae%x00%cn%x00%ce%x00%ai%x00%s")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return &api.AuthorScanResult{Commits: []api.MismatchedCommit{}, TotalCommits: 0, MatchCount: 0}, nil
	}

	lines := strings.Split(trimmed, "\n")
	var mismatched []api.MismatchedCommit

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\x00", 8)
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

		matchedAliases, ok := emailIndex[authorEmail]
		if !ok {
			continue
		}

		var match *api.MismatchedCommit
		for _, a := range matchedAliases {
			if authorName == a.name && authorEmail == a.email {
				if authorName == a.canonicalName && authorEmail == a.canonicalEmail {
					match = nil
					break
				}
				match = &api.MismatchedCommit{
					Hash: hash, ShortHash: shortHash, Message: message,
					AuthorName: authorName, AuthorEmail: authorEmail,
					CommitterName: committerName, CommitterEmail: committerEmail,
					Date: dateStr, MatchType: "exact",
					TargetName: a.canonicalName, TargetEmail: a.canonicalEmail,
				}
				break
			}
		}
		if match == nil {
			for _, a := range matchedAliases {
				if authorName == a.canonicalName && authorEmail == a.canonicalEmail {
					match = nil
					break
				}
				match = &api.MismatchedCommit{
					Hash: hash, ShortHash: shortHash, Message: message,
					AuthorName: authorName, AuthorEmail: authorEmail,
					CommitterName: committerName, CommitterEmail: committerEmail,
					Date: dateStr, MatchType: "email_only",
					TargetName: a.canonicalName, TargetEmail: a.canonicalEmail,
				}
				break
			}
		}
		if match != nil {
			mismatched = append(mismatched, *match)
		}
	}
	if mismatched == nil {
		mismatched = []api.MismatchedCommit{}
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

func cleanupBackupRefs(repoPath string) {
	cmd := runGit(repoPath, "for-each-ref", "--format=%(refname)", "refs/original/")
	refsOutput, _ := cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			runGit(repoPath, "update-ref", "-d", ref).Run()
		}
	}
}

func (s *AuthorService) fixAuthor(repoPath string, commitHashes []string, pushRemote string, taskID string) error {
	tm := GlobalTaskManager
	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
	}

	aliases, err := loadAllAliases()
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
	for _, a := range aliases {
		if a.name == a.canonicalName && a.email == a.canonicalEmail {
			continue
		}
		mappings = append(mappings, aliasMapping{
			oldName:  a.name,
			oldEmail: a.email,
			newName:  a.canonicalName,
			newEmail: a.canonicalEmail,
		})
	}

	if len(mappings) == 0 {
		appendLog("没有别名映射，无需修复")
		return nil
	}

	appendLog(fmt.Sprintf("共 %d 个别名映射，开始处理...", len(mappings)))

	var filterParts []string
	for _, m := range mappings {
		oE := shellEscape(m.oldEmail)
		oN := shellEscape(m.oldName)
		nN := shellEscape(m.newName)
		nE := shellEscape(m.newEmail)
		filterParts = append(filterParts,
			fmt.Sprintf(`if [ "$GIT_AUTHOR_EMAIL" = %s ] && [ "$GIT_AUTHOR_NAME" = %s ]; then export GIT_AUTHOR_NAME=%s; export GIT_AUTHOR_EMAIL=%s; fi`, oE, oN, nN, nE),
			fmt.Sprintf(`if [ "$GIT_AUTHOR_EMAIL" = %s ]; then export GIT_AUTHOR_NAME=%s; export GIT_AUTHOR_EMAIL=%s; fi`, oE, nN, nE),
			fmt.Sprintf(`if [ "$GIT_COMMITTER_EMAIL" = %s ] && [ "$GIT_COMMITTER_NAME" = %s ]; then export GIT_COMMITTER_NAME=%s; export GIT_COMMITTER_EMAIL=%s; fi`, oE, oN, nN, nE),
			fmt.Sprintf(`if [ "$GIT_COMMITTER_EMAIL" = %s ]; then export GIT_COMMITTER_NAME=%s; export GIT_COMMITTER_EMAIL=%s; fi`, oE, nN, nE),
		)
	}

	innerFilter := strings.Join(filterParts, "\n")

	var envFilter string
	if len(commitHashes) > 0 {
		appendLog(fmt.Sprintf("选择性修复模式: 仅修改 %d 个提交", len(commitHashes)))
		var hashConditions []string
		for _, h := range commitHashes {
			hashConditions = append(hashConditions, fmt.Sprintf(`[ "$GIT_COMMIT" = %s ]`, shellEscape(h)))
		}
		hashGuard := strings.Join(hashConditions, " || ")
		envFilter = fmt.Sprintf(`if %s; then
%s
fi`, hashGuard, innerFilter)
	} else {
		envFilter = innerFilter
	}

	appendLog("清理 backup refs...")
	cleanupBackupRefs(repoPath)

	appendLog("暂存工作区变更...")
	stashCmd := runGit(repoPath, "stash", "--include-untracked")
	stashOutput, stashErr := stashCmd.CombinedOutput()
	hasStash := stashErr == nil && !strings.Contains(string(stashOutput), "No local changes to save")

	appendLog("执行 filter-branch 重写作者信息...")
	gitEnv := append(os.Environ(), "FILTER_BRANCH_SQUELCH_WARNING=1")
	cmd := runGitWithEnv(repoPath, gitEnv,
		"filter-branch", "--force", "--env-filter", envFilter, "--tag-name-filter", "cat", "--", "--all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if hasStash {
			runGit(repoPath, "stash", "pop").CombinedOutput()
		}
		return fmt.Errorf("filter-branch failed: %w, output: %s", err, string(output))
	}
	appendLog("filter-branch 完成")

	if hasStash {
		appendLog("恢复工作区变更...")
		if popOut, popErr := runGit(repoPath, "stash", "pop").CombinedOutput(); popErr != nil {
			appendLog("stash pop 警告: " + string(popOut))
		}
	}

	appendLog("清理 refs/original/ ...")
	cleanupBackupRefs(repoPath)

	appendLog("清理 reflog...")
	runGit(repoPath, "reflog", "expire", "--expire=now", "--all").Run()

	appendLog("执行 gc...")
	runGit(repoPath, "gc", "--prune=now").CombinedOutput()

	if pushRemote != "" {
		appendLog("推送 " + pushRemote + " (force)...")
		if out, err := runGit(repoPath, "push", "--force", pushRemote, "--all").CombinedOutput(); err != nil {
			appendLog("push all 警告: " + string(out))
		} else {
			appendLog("push --all 完成")
		}
		if out, err := runGit(repoPath, "push", "--force", pushRemote, "--tags").CombinedOutput(); err != nil {
			appendLog("push tags 警告: " + string(out))
		} else {
			appendLog("push --tags 完成")
		}
	}

	appendLog("作者修复完成！")
	return nil
}

func (s *AuthorService) GetEffectiveAuthor(repoID uint) (name string, email string, err error) {
	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return "", "", fmt.Errorf("repo not found: %w", err)
	}
	if repo.AuthorIdentityID != nil {
		identity, err := db.NewAuthorIdentityDAO().FindByID(*repo.AuthorIdentityID)
		if err == nil {
			return identity.CanonicalName, identity.CanonicalEmail, nil
		}
	}
	identity, err := db.NewAuthorIdentityDAO().GetDefault()
	if err == nil {
		return identity.CanonicalName, identity.CanonicalEmail, nil
	}
	return "", "", nil
}

func StartAuthorFixTask(repoID uint, repoPath string, commitHashes []string, pushRemote string) (string, error) {
	authorFixMu.Lock()
	defer authorFixMu.Unlock()

	running := false
	GlobalTaskManager.tasks.Range(func(key, value interface{}) bool {
		t := value.(*Task)
		if t.Status == "running" {
			running = true
			return false
		}
		return true
	})
	if running {
		return "", fmt.Errorf("已有修复任务在运行，请等待完成后再试")
	}

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

func IdentityToDTO(id *po.AuthorIdentity) api.AuthorIdentityDTO {
	var aliases []api.AliasEntry
	if id.AliasesJSON != "" {
		_ = json.Unmarshal([]byte(id.AliasesJSON), &aliases)
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
