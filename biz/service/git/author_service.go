package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	servicePkg "github.com/yi-nology/git-manage-service/pkg/service"
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
	repo, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return nil, err
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
	repo, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return err
	}
	if clear {
		repo.AuthorIdentityID = nil
	} else {
		repo.AuthorIdentityID = identityID
	}
	return db.NewRepoDAO().Save(repo)
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
