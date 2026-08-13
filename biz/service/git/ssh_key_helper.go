package git

import (
	"net"

	"github.com/yi-nology/git-platform-sdk/pkg/credential"
	ssh2 "golang.org/x/crypto/ssh"
)

// SSHKeyHelper provides SSH key processing for git operations.
// It delegates to the SDK's credential.SSHKeyHelper.
type SSHKeyHelper struct {
	inner *credential.SSHKeyHelper
}

// NewSSHKeyHelper creates a new SSHKeyHelper instance.
func NewSSHKeyHelper() *SSHKeyHelper {
	return &SSHKeyHelper{inner: credential.NewSSHKeyHelper()}
}

// ProcessPrivateKey processes a private key, handling passphrase decryption.
func (h *SSHKeyHelper) ProcessPrivateKey(privateKey, passphrase string) (string, error) {
	return h.inner.ProcessPrivateKey(privateKey, passphrase)
}

// CreateTempKeyFile creates a temporary key file with 0600 permissions.
func (h *SSHKeyHelper) CreateTempKeyFile(keyContent string) (string, error) {
	return h.inner.CreateTempKeyFile(keyContent)
}

// BuildSSHCommand builds a GIT_SSH_COMMAND value with host key checking
// disabled. Suitable for CI/server environments where known_hosts cannot be
// populated.
//
// Note: git-platform-sdk v0.35+ made secure checking the default, so the SDK's
// BuildSSHCommand is now secure. This wrapper preserves the historical
// insecure behavior this codebase relied on by delegating to BuildSSHCommandInsecure.
func (h *SSHKeyHelper) BuildSSHCommand(keyPath string) string {
	return h.inner.BuildSSHCommandInsecure(keyPath)
}

// BuildSecureSSHCommand builds a GIT_SSH_COMMAND with strict host key
// verification enabled.
func (h *SSHKeyHelper) BuildSecureSSHCommand(keyPath string) string {
	return h.inner.BuildSSHCommand(keyPath)
}

// CleanupTempFile removes a temporary file.
func (h *SSHKeyHelper) CleanupTempFile(filePath string) {
	h.inner.CleanupTempFile(filePath)
}

// DetectKeyType detects the SSH key algorithm type.
func (h *SSHKeyHelper) DetectKeyType(privateKey, passphrase string) string {
	return h.inner.DetectKeyType(privateKey, passphrase)
}

// ExtractPublicKeyFromPrivateKey extracts the public key from a private key.
func (h *SSHKeyHelper) ExtractPublicKeyFromPrivateKey(privateKey, passphrase string) (string, error) {
	return h.inner.ExtractPublicKeyFromPrivateKey(privateKey, passphrase)
}

// AddHostKey registers a known host key.
func (h *SSHKeyHelper) AddHostKey(host string, key ssh2.PublicKey) {
	h.inner.AddHostKey(host, key)
}

// GetHostKeyCallback returns a host key callback for SSH connections.
func (h *SSHKeyHelper) GetHostKeyCallback() ssh2.HostKeyCallback {
	return h.inner.GetHostKeyCallback()
}

// HostKeyCallbackFunc is a convenience type for host key callbacks.
type HostKeyCallbackFunc = func(hostname string, remote net.Addr, key ssh2.PublicKey) error
