package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func (s *GitService) GetBranchSyncStatus(path, branch, upstream string) (int, int, error) {
	if upstream == "" {
		return 0, 0, nil
	}
	r, err := s.openRepo(path)
	if err != nil {
		return 0, 0, err
	}

	hBranch, err := r.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		hBranch, err = r.ResolveRevision(plumbing.Revision("refs/heads/" + branch))
		if err != nil {
			return 0, 0, nil
		}
	}

	hUpstream, err := r.ResolveRevision(plumbing.Revision(upstream))
	if err != nil {
		hUpstream, err = r.ResolveRevision(plumbing.Revision("refs/remotes/" + upstream))
		if err != nil {
			return 0, 0, nil
		}
	}

	cBranch, err := r.CommitObject(*hBranch)
	if err != nil {
		return 0, 0, err
	}
	cUpstream, err := r.CommitObject(*hUpstream)
	if err != nil {
		return 0, 0, err
	}

	bases, err := cBranch.MergeBase(cUpstream)
	if err != nil || len(bases) == 0 {
		return 0, 0, nil
	}
	base := bases[0]

	ahead := 0
	iter, err := r.Log(&git.LogOptions{From: *hBranch})
	if err == nil {
		_ = iter.ForEach(func(c *object.Commit) error {
			if c.Hash == base.Hash {
				return fmt.Errorf("stop")
			}
			ahead++
			return nil
		})
	}

	behind := 0
	iter, err = r.Log(&git.LogOptions{From: *hUpstream})
	if err == nil {
		_ = iter.ForEach(func(c *object.Commit) error {
			if c.Hash == base.Hash {
				return fmt.Errorf("stop")
			}
			behind++
			return nil
		})
	}

	return ahead, behind, nil
}

func (s *GitService) PushBranch(path, remote, branch string) error {
	return s.PushBranchWithAuth(path, remote, branch, nil)
}

func (s *GitService) PushBranchWithAuth(path, remote, branch string, auth transport.AuthMethod) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))

	if auth == nil {
		rem, err := r.Remote(remote)
		if err == nil {
			urls := rem.Config().URLs
			if len(urls) > 0 {
				auth = s.detectSSHAuth(urls[0])
			}
		}
	}

	pushOptions := &git.PushOptions{
		RemoteName: remote,
		RefSpecs:   []config.RefSpec{refSpec},
	}
	if auth != nil {
		pushOptions.Auth = auth
	}

	err = r.Push(pushOptions)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) PullBranch(path, remote, branch string) error {
	return s.PullBranchWithAuth(path, remote, branch, nil)
}

func (s *GitService) PullBranchWithAuth(path, remote, branch string, auth transport.AuthMethod) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	if auth == nil {
		rem, err := r.Remote(remote)
		if err == nil {
			urls := rem.Config().URLs
			if len(urls) > 0 {
				auth = s.detectSSHAuth(urls[0])
			}
		}
	}

	pullOptions := &git.PullOptions{
		RemoteName:    remote,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + branch),
	}
	if auth != nil {
		pullOptions.Auth = auth
	}

	err = w.Pull(pullOptions)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) UpdateBranchFastForward(path, remote, branch, remoteBranch string) error {
	return s.UpdateBranchFastForwardWithAuth(path, remote, branch, remoteBranch, nil)
}

func (s *GitService) UpdateBranchFastForwardWithAuth(path, remote, branch, remoteBranch string, auth transport.AuthMethod) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	rem, err := r.Remote(remote)
	if err != nil {
		return err
	}

	if auth == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", remoteBranch, branch))

	fetchOptions := &git.FetchOptions{
		RemoteName: remote,
		RefSpecs:   []config.RefSpec{refSpec},
	}
	if auth != nil {
		fetchOptions.Auth = auth
	}

	err = rem.Fetch(fetchOptions)

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) FetchAll(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		var auth transport.AuthMethod
		urls := remote.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}

		fetchOptions := &git.FetchOptions{
			RefSpecs: []config.RefSpec{
				config.RefSpec("+refs/heads/*:refs/remotes/" + remote.Config().Name + "/*"),
				config.RefSpec("+refs/tags/*:refs/tags/*"),
			},
		}
		if auth != nil {
			fetchOptions.Auth = auth
		}

		err := remote.Fetch(fetchOptions)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			_ = err
		}
	}
	return nil
}
