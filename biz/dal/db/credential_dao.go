package db

import (
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type CredentialDAO struct{ BaseDAO[po.Credential] }

func NewCredentialDAO() *CredentialDAO { return &CredentialDAO{} }

// FindAll 覆盖基类：按最近使用和更新时间排序
func (d *CredentialDAO) FindAll() ([]po.Credential, error) {
	var creds []po.Credential
	return creds, DB.Order("last_used_at DESC NULLS LAST, updated_at DESC").Find(&creds).Error
}

// FindNamesMap 批量查询凭证名称（轻量：不解密 Secret）
func (d *CredentialDAO) FindNamesMap(ids []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(ids) == 0 {
		return result, nil
	}
	type row struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	var rows []row
	if err := DB.Model(new(po.Credential)).Select("id, name").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.ID] = r.Name
	}
	return result, nil
}

// FindByName 根据名称查询
func (d *CredentialDAO) FindByName(name string) (*po.Credential, error) {
	var cred po.Credential
	return &cred, DB.Where("name = ?", name).First(&cred).Error
}

// FindByType 按类型查询
func (d *CredentialDAO) FindByType(credType string) ([]po.Credential, error) {
	var creds []po.Credential
	return creds, DB.Where("type = ?", credType).Order("last_used_at DESC NULLS LAST, updated_at DESC").Find(&creds).Error
}

// FindBySSHKeyID 查找引用指定 SSH 密钥的凭证
func (d *CredentialDAO) FindBySSHKeyID(sshKeyID uint) ([]po.Credential, error) {
	var creds []po.Credential
	return creds, DB.Where("type = ? AND ssh_key_id = ?", "ssh_key", sshKeyID).Find(&creds).Error
}

// UpdateLastUsed 更新最后使用时间
func (d *CredentialDAO) UpdateLastUsed(id uint) error {
	now := time.Now()
	return DB.Model(new(po.Credential)).Where("id = ?", id).Update("last_used_at", &now).Error
}

// FindMatchingURL 根据 URL 匹配凭证（通过 url_pattern 和协议类型）
func (d *CredentialDAO) FindMatchingURL(url string) (recommended []po.Credential, others []po.Credential, err error) {
	var all []po.Credential
	if err = DB.Order("last_used_at DESC NULLS LAST, updated_at DESC").Find(&all).Error; err != nil {
		return
	}
	isSSH := isSSHURL(url)
	for _, cred := range all {
		if cred.URLPattern != "" && matchURLPattern(cred.URLPattern, url) {
			recommended = append(recommended, cred)
			continue
		}
		if isSSH && cred.Type == "ssh_key" {
			recommended = append(recommended, cred)
		} else if !isSSH && (cred.Type == "http_basic" || cred.Type == "http_token") {
			recommended = append(recommended, cred)
		} else {
			others = append(others, cred)
		}
	}
	return
}

// isSSHURL 检测 URL 是否为 SSH 协议
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://") ||
		strings.Contains(url, "@") && !strings.HasPrefix(url, "http")
}

// matchURLPattern 简单的 URL 模式匹配（支持 * 通配符前缀）
func matchURLPattern(pattern, url string) bool {
	host := extractHost(url)
	if host == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:]
		return strings.HasSuffix(host, suffix) || host == pattern[2:]
	}
	return host == pattern
}

// extractHost 从 Git URL 中提取主机名
func extractHost(url string) string {
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url[4:], ":", 2)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	if strings.HasPrefix(url, "ssh://") {
		url = url[6:]
		if idx := strings.Index(url, "@"); idx >= 0 {
			url = url[idx+1:]
		}
		if idx := strings.IndexAny(url, ":/"); idx >= 0 {
			return url[:idx]
		}
		return url
	}
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		if idx := strings.IndexAny(url, ":/"); idx >= 0 {
			return url[:idx]
		}
		return url
	}
	return ""
}
