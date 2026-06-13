package mirror

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

func resolveAuth(mirror *po.Mirror) gitbackend.AuthConfig {
	if mirror.Credential == nil {
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	}

	cred := mirror.Credential
	switch cred.Type {
	case "ssh_key":
		return gitbackend.AuthConfig{
			Type:   gitbackend.AuthSSH,
			SSHKey: cred.SSHKeyPath,
		}
	case "http_basic":
		return gitbackend.AuthConfig{
			Type:     gitbackend.AuthHTTPBasic,
			Username: cred.Username,
			Password: cred.Secret,
		}
	case "http_token":
		return gitbackend.AuthConfig{
			Type:  gitbackend.AuthHTTPToken,
			Token: cred.Secret,
		}
	default:
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	}
}
