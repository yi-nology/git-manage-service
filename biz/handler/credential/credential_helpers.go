package credential

import (
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

var _ = time.Now

type sshKeyInfo struct {
	Name    string
	KeyType string
}

func buildSSHKeyMap(creds []po.Credential) map[uint]sshKeyInfo {
	m := make(map[uint]sshKeyInfo)
	sshKeyDAO := db.NewSSHKeyDAO()
	for _, cred := range creds {
		if cred.SSHKeyID > 0 {
			if _, ok := m[cred.SSHKeyID]; !ok {
				if key, err := sshKeyDAO.FindByID(cred.SSHKeyID); err == nil {
					m[cred.SSHKeyID] = sshKeyInfo{Name: key.Name, KeyType: key.KeyType}
				}
			}
		}
	}
	return m
}

func toCredentialDTO(cred *po.Credential, sshKeyMap map[uint]sshKeyInfo) api.CredentialDTO {
	dto := api.CredentialDTO{
		ID:          cred.ID,
		Name:        cred.Name,
		Type:        cred.Type,
		Description: cred.Description,
		SSHKeyID:    cred.SSHKeyID,
		SSHKeyPath:  cred.SSHKeyPath,
		Username:    cred.Username,
		HasSecret:   cred.Secret != "",
		URLPattern:  cred.URLPattern,
		LastUsedAt:  cred.LastUsedAt,
		CreatedAt:   cred.CreatedAt,
		UpdatedAt:   cred.UpdatedAt,
	}
	if cred.SSHKeyID > 0 {
		if sshKeyMap != nil {
			if info, ok := sshKeyMap[cred.SSHKeyID]; ok {
				dto.SSHKeyName = info.Name
				dto.SSHKeyType = info.KeyType
			}
		} else {
			sshKeyDAO := db.NewSSHKeyDAO()
			if key, err := sshKeyDAO.FindByID(cred.SSHKeyID); err == nil {
				dto.SSHKeyName = key.Name
				dto.SSHKeyType = key.KeyType
			}
		}
	}
	if dto.SSHKeyPath != "" {
		parts := splitPath(dto.SSHKeyPath)
		if len(parts) > 0 {
			dto.SSHKeyPath = ".../" + parts[len(parts)-1]
		}
	}
	return dto
}

func splitPath(p string) []string {
	var parts []string
	for _, s := range []string{"/", "\\"} {
		if len(p) > 0 {
			for _, part := range splitByDelimiter(p, s) {
				if part != "" {
					parts = append(parts, part)
				}
			}
			if len(parts) > 0 {
				return parts
			}
		}
	}
	return []string{p}
}

func splitByDelimiter(s, delim string) []string {
	var result []string
	for {
		i := indexString(s, delim)
		if i < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:i])
		s = s[i+len(delim):]
	}
	return result
}

func indexString(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func parseID(c *app.RequestContext) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	return uint(id), err
}
