package git

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

func parsePushOptions(options []string) (force bool, mirror bool, insecure bool) {
	for _, o := range options {
		if o == "-f" || o == "--force" {
			force = true
		} else if o == "--mirror" {
			mirror = true
		} else if o == "--insecure" {
			insecure = true
		}
	}
	return
}

func (s *GitService) Push(path, targetRemote, sourceHash, targetBranch string, options []string, progress io.Writer, skipTLS ...bool) error {
	force, mirror, _ := parsePushOptions(options)
	insecure := len(skipTLS) > 0 && skipTLS[0]

	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)

	_, err := s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath:        path,
		Remote:          targetRemote,
		RefSpecs:        []string{refSpec},
		Force:           force,
		Mirror:          mirror,
		InsecureSkipTLS: insecure,
		Progress:        progress,
	})
	return err
}

func (s *GitService) PushWithAuth(path, targetRemoteURL, sourceHash, targetBranch, authType, authKey, authSecret string, options []string, progress io.Writer, skipTLS ...bool) error {
	auth := s.buildSDKAuth(authType, authKey, authSecret)

	return s.PushWithSDKAuth(path, targetRemoteURL, sourceHash, targetBranch, auth, options, progress, skipTLS...)
}

func (s *GitService) PushWithAuthMethod(path, targetRemoteURL, sourceHash, targetBranch string, authMethod transport.AuthMethod, options []string, progress io.Writer, skipTLS ...bool) error {
	sdkAuth := s.ConvertTransportAuth(authMethod)
	return s.PushWithSDKAuth(path, targetRemoteURL, sourceHash, targetBranch, sdkAuth, options, progress, skipTLS...)
}

func (s *GitService) PushWithSDKAuth(path, targetRemoteURL, sourceHash, targetBranch string, auth gitbackend.AuthConfig, options []string, progress io.Writer, skipTLS ...bool) error {
	force, mirror, _ := parsePushOptions(options)
	insecure := len(skipTLS) > 0 && skipTLS[0]

	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)

	_, err := s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath:        path,
		Remote:          targetRemoteURL,
		RefSpecs:        []string{refSpec},
		Force:           force,
		Mirror:          mirror,
		InsecureSkipTLS: insecure,
		Auth:            auth,
		Progress:        progress,
	})
	return err
}

func (s *GitService) PushCurrent(path string) error {
	_, err := s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath: path,
		Remote:   "origin",
	})
	return err
}

func (s *GitService) buildSDKAuth(authType, authKey, authSecret string) gitbackend.AuthConfig {
	switch authType {
	case "http":
		if authKey != "" {
			return gitbackend.AuthConfig{
				Type:     gitbackend.AuthHTTPBasic,
				Username: authKey,
				Password: authSecret,
			}
		}
	case "ssh":
		if authKey != "" {
			return gitbackend.AuthConfig{
				Type:   gitbackend.AuthSSH,
				SSHKey: authKey,
			}
		}
	}
	return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
}

func (s *GitService) ConvertTransportAuth(authMethod transport.AuthMethod) gitbackend.AuthConfig {
	if authMethod == nil {
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	}

	switch a := authMethod.(type) {
	case *ssh.PublicKeys:
		return gitbackend.AuthConfig{
			Type:   gitbackend.AuthSSH,
			SSHKey: a.User,
		}
	case *http.BasicAuth:
		return gitbackend.AuthConfig{
			Type:     gitbackend.AuthHTTPBasic,
			Username: a.Username,
			Password: a.Password,
		}
	default:
		return gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	}
}
