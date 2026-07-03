package git

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yi-nology/git-platform-sdk/gitbackend"

	conf "github.com/yi-nology/git-manage-service/pkg/configs"
)

// TestRemoteConnectionWithDBKey 使用数据库密钥测试远程连接
func (s *GitService) TestRemoteConnectionWithDBKey(url, privateKey, passphrase string, skipTLS ...bool) error {
	// 优先使用原生 git 命令测试连接（最可靠）
	if err := s.testConnectionWithGitCommand(url, privateKey, passphrase); err == nil {
		return nil
	}

	// 回退：通过 SDK backend 使用密钥内容测试
	helper := NewSSHKeyHelper()
	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to process private key: %v", err)
	}

	auth := gitbackend.NewSSHKeyContentAuth(keyContent, passphrase)
	auth.InsecureSkipTLS = len(skipTLS) > 0 && skipTLS[0]

	return s.backend.TestConnection(context.Background(), url, auth)
}

// testConnectionWithGitCommand 使用原生 git 命令测试连接（更可靠）
func (s *GitService) testConnectionWithGitCommand(url, privateKey, passphrase string) error {
	helper := NewSSHKeyHelper()

	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		return fmt.Errorf("process key failed: %v", err)
	}

	tmpFile, err := helper.CreateTempKeyFile(keyContent)
	if err != nil {
		return err
	}
	defer helper.CleanupTempFile(tmpFile)

	sshCmd := helper.BuildSSHCommand(tmpFile)

	cmd := exec.Command("git", "ls-remote", "--heads", url)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git ls-remote failed: %v, output: %s", err, string(output))
	}

	return nil
}

// TestRemoteConnectionWithLocalKey 使用本地SSH密钥文件测试远程连接
func (s *GitService) TestRemoteConnectionWithLocalKey(url, keyPath, passphrase string, skipTLS ...bool) error {
	auth := gitbackend.NewSSHKeyFileAuth(keyPath, passphrase)
	auth.InsecureSkipTLS = len(skipTLS) > 0 && skipTLS[0]
	return s.backend.TestConnection(context.Background(), url, auth)
}

// TestRemoteConnectionWithHTTP 使用HTTP认证测试远程连接
func (s *GitService) TestRemoteConnectionWithHTTP(url, username, password string, skipTLS ...bool) error {
	auth := gitbackend.NewHTTPBasicAuth(username, password)
	auth.InsecureSkipTLS = len(skipTLS) > 0 && skipTLS[0]
	return s.backend.TestConnection(context.Background(), url, auth)
}

// detectSSHAuth auto-detects an SSH key from common paths and returns an SDK
// gitbackend.AuthConfig. Returns AuthNone for non-SSH URLs or when no key is found.
func (s *GitService) detectSSHAuth(urlStr string) gitbackend.AuthConfig {
	if strings.HasPrefix(urlStr, "https://") || strings.HasPrefix(urlStr, "http://") {
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	}

	if !strings.HasPrefix(urlStr, "git@") && !strings.HasPrefix(urlStr, "ssh://") {
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	}

	if conf.DebugMode {
		log.Printf("[DEBUG] detectSSHAuth for %s", urlStr)
	}

	// Try common key paths (unencrypted)
	home, err := os.UserHomeDir()
	if err == nil {
		keyPaths := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_ecdsa"),
		}

		for _, path := range keyPaths {
			if _, err := os.Stat(path); err == nil {
				if conf.DebugMode {
					log.Printf("[DEBUG] Using SSH Key: %s", path)
				}
				return gitbackend.NewSSHKeyFileAuth(path, "")
			}
		}
	}

	// No explicit key found: return AuthNone so the native git backend falls
	// back to the SSH agent / system SSH config automatically.
	if conf.DebugMode {
		log.Printf("[DEBUG] No SSH key file found, relying on SSH agent/default config")
	}
	return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
}
