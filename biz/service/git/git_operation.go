package git

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type channelWriter struct {
	ch chan string
}

func (w *channelWriter) Write(p []byte) (n int, err error) {
	if w.ch != nil {
		w.ch <- string(p)
	}
	return len(p), nil
}

func (s *GitService) openRepo(path string) (*git.Repository, error) {
	log.Printf("[DEBUG] Opening repository at: %s", path)
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Printf("[ERROR] Failed to open repository at %s: %v", path, err)
		return nil, fmt.Errorf("failed to open repository at %s: %v", path, err)
	}
	log.Printf("[DEBUG] Repository opened successfully: %s", path)
	return r, nil
}

func (s *GitService) IsGitRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}

func (s *GitService) Fetch(path, remote string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	var auth transport.AuthMethod
	rem, err := r.Remote(remote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	fetchOptions := &git.FetchOptions{
		RemoteName: remote,
		Progress:   progress,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/remotes/" + remote + "/*"),
			config.RefSpec("+refs/tags/*:refs/tags/*"),
		},
	}
	if auth != nil {
		fetchOptions.Auth = auth
	}

	err = r.Fetch(fetchOptions)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) FetchWithAuth(path, remoteURL, authType, authKey, authSecret string, progress io.Writer, extraArgs ...string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})

	err = remote.Fetch(&git.FetchOptions{
		Auth:       auth,
		RemoteName: "origin",
		Progress:   progress,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) FetchWithAuthMethod(path, remoteURL string, auth transport.AuthMethod, progress io.Writer, extraArgs ...string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})

	err = remote.Fetch(&git.FetchOptions{
		Auth:       auth,
		RemoteName: "origin",
		Progress:   progress,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) Clone(remoteURL, localPath, authType, authKey, authSecret string) error {
	return s.CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret, nil)
}

func (s *GitService) CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret string, progressChan chan string) error {
	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	if auth == nil {
		auth = s.detectSSHAuth(remoteURL)
	}

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	_, err = git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      remoteURL,
		Auth:     auth,
		Progress: progress,
	})
	return err
}

func (s *GitService) CloneWithAuthMethod(remoteURL, localPath string, auth transport.AuthMethod, progressChan chan string) error {
	if auth == nil {
		auth = s.detectSSHAuth(remoteURL)
	}

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	_, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      remoteURL,
		Auth:     auth,
		Progress: progress,
	})
	return err
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remote, branch))
	ref, err := r.Reference(refName, true)
	if err != nil {
		return "", fmt.Errorf("remote branch %s/%s not found: %v", remote, branch, err)
	}
	return ref.Hash().String(), nil
}

func (s *GitService) IsAncestor(path, ancestor, descendant string) (bool, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return false, err
	}

	h1 := plumbing.NewHash(ancestor)
	h2 := plumbing.NewHash(descendant)

	c1, err := r.CommitObject(h1)
	if err != nil {
		return false, err
	}
	c2, err := r.CommitObject(h2)
	if err != nil {
		return false, err
	}

	return c1.IsAncestor(c2)
}

func (s *GitService) GetBranches(path string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	iter, err := r.References()
	if err != nil {
		return nil, err
	}
	var branches []string
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() || ref.Name().IsRemote() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func (s *GitService) GetCommits(path, branch, since, until string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return "", err
	}

	cIter, err := r.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	forEachErr := cIter.ForEach(func(c *object.Commit) error {
		line := fmt.Sprintf("%s|%s|%s|%s|%s\n",
			c.Hash.String(),
			c.Author.Name,
			c.Author.Email,
			c.Author.When.Format("2006-01-02 15:04:05 -0700"),
			strings.TrimSpace(strings.Split(c.Message, "\n")[0]),
		)
		sb.WriteString(line)
		return nil
	})
	if forEachErr != nil {
		return "", forEachErr
	}

	return sb.String(), nil
}

func (s *GitService) GetRepoFiles(path, branch string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []string
	_ = tree.Files().ForEach(func(f *object.File) error {
		files = append(files, f.Name)
		return nil
	})
	return files, nil
}

func (s *GitService) BlameFile(path, branch, file string) (*git.BlameResult, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return nil, err
	}

	return git.Blame(commit, file)
}

func (s *GitService) GetCommit(path, hashStr string) (*object.Commit, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	return r.CommitObject(plumbing.NewHash(hashStr))
}

func (s *GitService) resolveCommit(r *git.Repository, rev string) (*object.Commit, error) {
	hash, err := r.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		if !strings.HasPrefix(rev, "refs/") {
			h, err2 := r.ResolveRevision(plumbing.Revision("refs/heads/" + rev))
			if err2 == nil {
				hash = h
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return r.CommitObject(*hash)
}

func (s *GitService) resolveCommitPair(r *git.Repository, base, target string) (*object.Commit, *object.Commit, error) {
	cBase, err := s.resolveCommit(r, base)
	if err != nil {
		return nil, nil, err
	}
	cTarget, err := s.resolveCommit(r, target)
	if err != nil {
		return nil, nil, err
	}
	return cBase, cTarget, nil
}

func (s *GitService) ResolveRevision(path, rev string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	hash, err := r.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (s *GitService) GetHeadBranch(path string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	head, err := r.Head()
	if err != nil {
		return "", err
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return head.Hash().String(), nil
}
