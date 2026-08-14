package credential

import (
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	credential "github.com/yi-nology/git-manage-service/biz/model/credential"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
)

type CredentialService struct {
	credDAO *db.CredentialDAO
	repoDAO *db.RepoDAO
	pcDAO   *db.ProviderConfigDAO
	sshDAO  *db.SSHKeyDAO
	gitSvc  *git.GitService
}

func NewCredentialService() *CredentialService {
	return &CredentialService{
		credDAO: db.NewCredentialDAO(),
		repoDAO: db.NewRepoDAO(),
		pcDAO:   db.NewProviderConfigDAO(),
		sshDAO:  db.NewSSHKeyDAO(),
		gitSvc:  git.NewGitService(),
	}
}

func (s *CredentialService) List(req *credential.ListCredentialsRequest) ([]*credential.CredentialInfo, error) {
	// Push the type filter to SQL when set (avoids loading + decrypting all
	// credentials just to discard most of them).
	var creds []po.Credential
	var err error
	if req.Type != "" {
		creds, err = s.credDAO.FindByType(req.Type)
	} else {
		creds, err = s.credDAO.FindAll()
	}
	if err != nil {
		return nil, err
	}

	sshKeyMap := s.buildSSHKeyMap(creds)
	result := make([]*credential.CredentialInfo, 0, len(creds))

	for _, cred := range creds {
		// Purpose is an orthogonal dimension to type (http_basic/http_token
		// satisfy both purposes) — always apply it.
		if req.Purpose != "" {
			hasPurpose := false
			switch req.Purpose {
			case "git_remote":
				hasPurpose = cred.Type == "ssh_key" || cred.Type == "http_basic" || cred.Type == "http_token"
			case "platform_api":
				hasPurpose = cred.Type == "platform_token" || cred.Type == "http_basic" || cred.Type == "http_token"
			}
			if !hasPurpose {
				continue
			}
		}

		dto := s.toCredentialDTO(&cred, sshKeyMap)

		if req.Status != "" && dto.Status != req.Status {
			continue
		}

		result = append(result, dto)
	}

	return result, nil
}

func (s *CredentialService) Create(req *credential.CreateCredentialRequest) (*credential.CredentialInfo, error) {
	if err := validateCredential(req.Name, req.Type, req.SshKeyId, req.SshKeyPath); err != nil {
		return nil, err
	}

	if req.SshKeyId > 0 {
		if _, err := s.sshDAO.FindByID(uint(req.SshKeyId)); err != nil {
			return nil, err
		}
	}

	cred := &po.Credential{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		SSHKeyID:    uint(req.SshKeyId),
		SSHKeyPath:  req.SshKeyPath,
		Username:    req.Username,
		Secret:      req.Secret,
		URLPattern:  req.UrlPattern,
	}

	if err := s.credDAO.Create(cred); err != nil {
		return nil, err
	}

	return s.toCredentialDTO(cred, nil), nil
}

func (s *CredentialService) Get(id uint) (*credential.CredentialInfo, error) {
	cred, err := s.credDAO.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.toCredentialDTO(cred, nil), nil
}

func (s *CredentialService) Update(id uint, req *credential.UpdateCredentialRequest) (*credential.CredentialInfo, error) {
	cred, err := s.credDAO.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		cred.Name = req.Name
	}
	if req.Description != "" {
		cred.Description = req.Description
	}
	if req.SshKeyId > 0 {
		if _, err := s.sshDAO.FindByID(uint(req.SshKeyId)); err != nil {
			return nil, err
		}
		cred.SSHKeyID = uint(req.SshKeyId)
	}
	if req.SshKeyPath != "" {
		cred.SSHKeyPath = req.SshKeyPath
	}
	if req.Username != "" {
		cred.Username = req.Username
	}
	if req.Secret != "" {
		cred.Secret = req.Secret
	}
	if req.UrlPattern != "" {
		cred.URLPattern = req.UrlPattern
	}

	if err := s.credDAO.Save(cred); err != nil {
		return nil, err
	}

	return s.toCredentialDTO(cred, nil), nil
}

func (s *CredentialService) Delete(id uint) error {
	if _, err := s.credDAO.FindByID(id); err != nil {
		return err
	}

	repos, err := s.repoDAO.FindAll()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if repo.DefaultCredentialID == id {
			return ErrCredentialInUse
		}
		if repo.RemoteCredentials != nil {
			for _, credID := range repo.RemoteCredentials {
				if credID == id {
					return ErrCredentialInUse
				}
			}
		}
	}

	configs, err := s.pcDAO.FindAll()
	if err != nil {
		return err
	}
	for _, cfg := range configs {
		if cfg.CredentialID == id {
			return ErrCredentialInUse
		}
	}

	return s.credDAO.Delete(id)
}

func (s *CredentialService) GetUsages(id uint) (*credential.GetCredentialUsagesResponse, error) {
	if _, err := s.credDAO.FindByID(id); err != nil {
		return nil, err
	}

	repos, err := s.repoDAO.FindAll()
	if err != nil {
		return nil, err
	}
	var repoUsages []*credential.CredentialRepoUsage
	for _, repo := range repos {
		if repo.DefaultCredentialID == id {
			repoUsages = append(repoUsages, &credential.CredentialRepoUsage{
				RepoId:     uint64(repo.ID),
				RepoName:   repo.Name,
				RepoKey:    repo.Key,
				IsDefault:  true,
				RemoteName: "",
			})
		}
		if repo.RemoteCredentials != nil {
			for remoteName, credID := range repo.RemoteCredentials {
				if credID == id {
					repoUsages = append(repoUsages, &credential.CredentialRepoUsage{
						RepoId:     uint64(repo.ID),
						RepoName:   repo.Name,
						RepoKey:    repo.Key,
						IsDefault:  false,
						RemoteName: remoteName,
					})
				}
			}
		}
	}

	configs, err := s.pcDAO.FindAll()
	if err != nil {
		return nil, err
	}
	var providerUsages []*credential.CredentialProviderUsage
	for _, cfg := range configs {
		if cfg.CredentialID == id {
			providerUsages = append(providerUsages, &credential.CredentialProviderUsage{
				ProviderId:   uint64(cfg.ID),
				ProviderName: cfg.Name,
				Platform:     cfg.Platform,
			})
		}
	}

	return &credential.GetCredentialUsagesResponse{
		Repos:              repoUsages,
		Providers:          providerUsages,
		TotalRepoCount:     int32(len(repoUsages)),
		TotalProviderCount: int32(len(providerUsages)),
	}, nil
}

func (s *CredentialService) Rotate(id uint, req *credential.RotateCredentialRequest) error {
	cred, err := s.credDAO.FindByID(id)
	if err != nil {
		return err
	}

	oldSecret := cred.Secret
	cred.Secret = req.NewSecret

	if req.TestUrl != "" {
		var testErr error
		switch cred.Type {
		case "ssh_key":
			if cred.SSHKeyID > 0 {
				sshKey, err := s.sshDAO.FindByID(cred.SSHKeyID)
				if err != nil {
					return err
				}
				testErr = s.gitSvc.TestRemoteConnectionWithDBKey(req.TestUrl, sshKey.PrivateKey, sshKey.Passphrase)
			} else if cred.SSHKeyPath != "" {
				testErr = s.gitSvc.TestRemoteConnectionWithLocalKey(req.TestUrl, cred.SSHKeyPath, "")
			} else {
				testErr = s.gitSvc.TestRemoteConnection(req.TestUrl)
			}
		case "http_basic", "http_token":
			testErr = s.gitSvc.TestRemoteConnectionWithHTTP(req.TestUrl, cred.Username, req.NewSecret)
		default:
			testErr = s.gitSvc.TestRemoteConnection(req.TestUrl)
		}

		now := time.Now()
		cred.LastTestedAt = &now
		if testErr != nil {
			cred.LastTestOk = false
			cred.LastError = testErr.Error()
			_ = s.credDAO.Save(cred)
			return testErr
		}
		cred.LastTestOk = true
		cred.LastError = ""
	}

	now := time.Now()
	cred.RotatedAt = &now
	cred.LastUsedAt = &now

	if err := s.credDAO.Save(cred); err != nil {
		cred.Secret = oldSecret
		return err
	}

	return nil
}

func (s *CredentialService) Match(url string) (*credential.MatchCredentialResponse, error) {
	recommended, others, err := s.credDAO.FindMatchingURL(url)
	if err != nil {
		return nil, err
	}

	resp := &credential.MatchCredentialResponse{
		Recommended: make([]*credential.CredentialInfo, 0, len(recommended)),
		Others:      make([]*credential.CredentialInfo, 0, len(others)),
	}

	for _, cred := range recommended {
		resp.Recommended = append(resp.Recommended, s.toCredentialDTO(&cred, nil))
	}
	for _, cred := range others {
		resp.Others = append(resp.Others, s.toCredentialDTO(&cred, nil))
	}

	return resp, nil
}

func (s *CredentialService) TestConnection(id uint, url string) (bool, string, error) {
	cred, err := s.credDAO.FindByID(id)
	if err != nil {
		return false, "", err
	}

	gitSvc := s.gitSvc
	var testErr error

	switch cred.Type {
	case "ssh_key":
		if cred.SSHKeyID > 0 {
			sshKey, err := s.sshDAO.FindByID(cred.SSHKeyID)
			if err != nil {
				return false, "Failed to load SSH key", err
			}
			testErr = gitSvc.TestRemoteConnectionWithDBKey(url, sshKey.PrivateKey, sshKey.Passphrase)
		} else if cred.SSHKeyPath != "" {
			testErr = gitSvc.TestRemoteConnectionWithLocalKey(url, cred.SSHKeyPath, "")
		} else {
			testErr = gitSvc.TestRemoteConnection(url)
		}
	case "http_basic", "http_token":
		testErr = gitSvc.TestRemoteConnectionWithHTTP(url, cred.Username, cred.Secret)
	default:
		testErr = gitSvc.TestRemoteConnection(url)
	}

	now := time.Now()
	cred.LastTestedAt = &now
	if testErr != nil {
		cred.LastTestOk = false
		cred.LastError = testErr.Error()
		_ = s.credDAO.Save(cred)
		return false, testErr.Error(), nil
	}

	cred.LastTestOk = true
	cred.LastError = ""
	cred.LastUsedAt = &now
	_ = s.credDAO.Save(cred)
	return true, "Connection successful", nil
}

func (s *CredentialService) buildSSHKeyMap(creds []po.Credential) map[uint]sshKeyInfo {
	m := make(map[uint]sshKeyInfo)
	for _, cred := range creds {
		if cred.SSHKeyID > 0 {
			if _, ok := m[cred.SSHKeyID]; !ok {
				if key, err := s.sshDAO.FindByID(cred.SSHKeyID); err == nil {
					m[cred.SSHKeyID] = sshKeyInfo{Name: key.Name, KeyType: key.KeyType}
				}
			}
		}
	}
	return m
}

func (s *CredentialService) toCredentialDTO(cred *po.Credential, sshKeyMap map[uint]sshKeyInfo) *credential.CredentialInfo {
	dto := &credential.CredentialInfo{
		Id:           uint64(cred.ID),
		Name:         cred.Name,
		Type:         cred.Type,
		Description:  cred.Description,
		SshKeyId:     uint64(cred.SSHKeyID),
		SshKeyPath:   cred.SSHKeyPath,
		Username:     cred.Username,
		HasSecret:    cred.Secret != "",
		UrlPattern:   cred.URLPattern,
		Status:       s.calculateCredentialStatus(cred),
		LastUsedAt:   formatTime(cred.LastUsedAt),
		LastTestedAt: formatTime(cred.LastTestedAt),
		LastTestOk:   cred.LastTestOk,
		LastError:    cred.LastError,
		RotatedAt:    formatTime(cred.RotatedAt),
		ExpiresAt:    formatTime(cred.ExpiresAt),
		CreatedAt:    formatTime(&cred.CreatedAt),
		UpdatedAt:    formatTime(&cred.UpdatedAt),
	}
	if cred.SSHKeyID > 0 {
		if sshKeyMap != nil {
			if info, ok := sshKeyMap[cred.SSHKeyID]; ok {
				dto.SshKeyName = info.Name
				dto.SshKeyType = info.KeyType
			}
		} else {
			if key, err := s.sshDAO.FindByID(cred.SSHKeyID); err == nil {
				dto.SshKeyName = key.Name
				dto.SshKeyType = key.KeyType
			}
		}
	}
	if dto.SshKeyPath != "" {
		base := dto.SshKeyPath
		if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
			base = base[i+1:]
		}
		dto.SshKeyPath = ".../" + base
	}
	return dto
}

func (s *CredentialService) calculateCredentialStatus(cred *po.Credential) string {
	if cred.LastTestedAt != nil && !cred.LastTestOk {
		return "invalid"
	}
	if cred.ExpiresAt != nil {
		daysUntilExpire := time.Until(*cred.ExpiresAt).Hours() / 24
		if daysUntilExpire <= 0 {
			return "expired"
		}
		if daysUntilExpire <= 7 {
			return "expiring"
		}
	}
	return "active"
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

type sshKeyInfo struct {
	Name    string
	KeyType string
}
