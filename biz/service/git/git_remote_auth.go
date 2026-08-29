package git

import (
	"fmt"

	"github.com/yi-nology/git-platform-sdk/gitbackend"

	"github.com/yi-nology/git-manage-service/biz/service/auth"
)

// RemoteAuth 为单个 remote 解析并准备好的认证材料。
type RemoteAuth struct {
	// Method 为非数据库密钥场景下解析出的认证配置（Type 为 AuthNone 表示无需认证）。
	Method gitbackend.AuthConfig
	// IsDBKey 表示命中数据库 SSH 密钥体系。
	IsDBKey bool
	// CredID 为该 remote 关联的凭证 ID；IsDBKey 且 CredID 为 0 表示未配置凭证。
	CredID uint
	// PrivateKey / Passphrase 仅在 HasDBKeyContent() 为 true 时有效。
	PrivateKey string
	Passphrase string
}

// HasDBKeyContent 报告是否已加载出可直接使用的数据库 SSH 私钥内容。
func (a RemoteAuth) HasDBKeyContent() bool {
	return a.IsDBKey && a.CredID > 0
}

// NoCredential 报告命中数据库密钥体系但该 remote 未配置任何凭证。
func (a RemoteAuth) NoCredential() bool {
	return a.IsDBKey && a.CredID == 0
}

// PrepareRemoteAuth 解析指定 remote 的认证：先解析凭证，若为数据库 SSH 密钥
// 则进一步加载私钥内容。既有 handler 调用均未配置旧版 RemoteAuths 与默认认证
// （传 nil 与空串），此处固定沿用该形态。
// 出错时返回形如 "failed to resolve auth: ..." 或 "failed to load SSH key: ..."
// 的错误，调用方按 "<remote>: <err>" 拼接后即为最终文案。
func PrepareRemoteAuth(
	authSvc *auth.AuthService,
	remoteCredentials map[string]uint,
	defaultCredentialID uint,
	remote string,
) (RemoteAuth, error) {
	authMethod, isDBKey, resolveErr := authSvc.ResolveCredentialForRemote(
		remoteCredentials,
		defaultCredentialID,
		nil,
		remote,
		"", "", "",
	)
	if resolveErr != nil {
		return RemoteAuth{}, fmt.Errorf("failed to resolve auth: %w", resolveErr)
	}
	if !isDBKey {
		return RemoteAuth{Method: authMethod}, nil
	}

	credID := auth.GetCredentialIDForRemote(remoteCredentials, defaultCredentialID, remote)
	if credID == 0 {
		return RemoteAuth{IsDBKey: true}, nil
	}

	privateKey, passphrase, keyErr := authSvc.GetCredentialKeyContent(credID)
	if keyErr != nil {
		return RemoteAuth{}, fmt.Errorf("failed to load SSH key: %w", keyErr)
	}
	return RemoteAuth{IsDBKey: true, CredID: credID, PrivateKey: privateKey, Passphrase: passphrase}, nil
}
