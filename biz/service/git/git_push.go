package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func parsePushOptions(options []string) *git.PushOptions {
	opts := &git.PushOptions{}
	opts.Options = make(map[string]string)

	for _, o := range options {
		if o == "-f" || o == "--force" {
			opts.Force = true
		} else if o == "--prune" {
			opts.Prune = true
		} else if strings.HasPrefix(o, "--push-option=") {
			kv := strings.TrimPrefix(o, "--push-option=")
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				opts.Options[parts[0]] = parts[1]
			}
		}
	}
	return opts
}

func (s *GitService) Push(path, targetRemote, sourceHash, targetBranch string, options []string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch))

	var auth transport.AuthMethod
	rem, err := r.Remote(targetRemote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	pushOpts := parsePushOptions(options)
	pushOpts.RemoteName = targetRemote
	pushOpts.RefSpecs = []config.RefSpec{refSpec}
	if auth != nil {
		pushOpts.Auth = auth
	}
	pushOpts.Progress = progress

	err = r.Push(pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) PushWithAuth(path, targetRemoteURL, sourceHash, targetBranch, authType, authKey, authSecret string, options []string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "anonymous",
		URLs: []string{targetRemoteURL},
	})

	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch))

	pushOpts := parsePushOptions(options)
	pushOpts.Auth = auth
	pushOpts.RefSpecs = []config.RefSpec{refSpec}
	pushOpts.Progress = progress

	err = remote.Push(pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) PushWithAuthMethod(path, targetRemoteURL, sourceHash, targetBranch string, auth transport.AuthMethod, options []string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "anonymous",
		URLs: []string{targetRemoteURL},
	})

	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch))

	pushOpts := parsePushOptions(options)
	pushOpts.Auth = auth
	pushOpts.RefSpecs = []config.RefSpec{refSpec}
	pushOpts.Progress = progress

	err = remote.Push(pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) PushCurrent(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	var auth transport.AuthMethod
	rem, err := r.Remote("origin")
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	err = r.Push(&git.PushOptions{
		Auth: auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}
