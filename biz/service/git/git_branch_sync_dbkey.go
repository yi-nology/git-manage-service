package git

import (
	"context"
	"fmt"

	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

// dbKeyAuth normalizes a (possibly passphrase-protected) private key and builds
// an SDK AuthConfig using key content. The native backend writes the content to
// a temp file and sets GIT_SSH_COMMAND automatically.
func (s *GitService) dbKeyAuth(privateKey, passphrase string) (gitbackend.AuthConfig, error) {
	helper := NewSSHKeyHelper()
	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}, fmt.Errorf("failed to process private key: %v", err)
	}
	return gitbackend.NewSSHKeyContentAuth(keyContent, ""), nil
}

func (s *GitService) PushBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	auth, err := s.dbKeyAuth(privateKey, passphrase)
	if err != nil {
		return err
	}

	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)

	_, err = s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath: path,
		Remote:   remote,
		RefSpecs: []string{refSpec},
		Auth:     auth,
	})
	return err
}

func (s *GitService) PullBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	auth, err := s.dbKeyAuth(privateKey, passphrase)
	if err != nil {
		return err
	}

	if err := s.backend.Pull(context.Background(), path, remote, branch, auth); err != nil {
		return fmt.Errorf("git pull failed: %v", err)
	}
	return nil
}

func (s *GitService) FetchBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	auth, err := s.dbKeyAuth(privateKey, passphrase)
	if err != nil {
		return err
	}

	_, err = s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath: path,
		Remote:   remote,
		Branches: []string{branch},
		Auth:     auth,
	})
	return err
}
