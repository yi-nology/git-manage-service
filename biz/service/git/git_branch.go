package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
)

// ListBranchesWithInfo returns detailed information for all branches
func (s *GitService) ListBranchesWithInfo(path string) ([]domain.BranchInfo, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	iter, err := r.References()
	if err != nil {
		return nil, err
	}

	headRef, err := r.Head()
	var headHash plumbing.Hash
	if err == nil {
		headHash = headRef.Hash()
	}

	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}

	var branches []domain.BranchInfo

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		isBranch := ref.Name().IsBranch()
		isRemote := ref.Name().IsRemote()

		if !isBranch && !isRemote {
			return nil
		}

		name := ref.Name().Short()
		hash := ref.Hash()

		b := domain.BranchInfo{
			Name: name,
			Hash: hash.String(),
		}

		if isBranch {
			b.Type = "local"
		} else if isRemote {
			b.Type = "remote"
		} else {
			return nil
		}

		if hash == headHash && ref.Name().IsBranch() {
			b.IsCurrent = true
		}

		commit, commitErr := r.CommitObject(hash)
		if commitErr == nil {
			b.Author = commit.Author.Name
			b.AuthorEmail = commit.Author.Email
			b.Date = commit.Author.When
			b.Message = strings.TrimSpace(strings.Split(commit.Message, "\n")[0])
		}

		if ref.Name().IsBranch() {
			if branchCfg, ok := cfg.Branches[name]; ok {
				if branchCfg.Remote != "" && branchCfg.Merge != "" {
					shortMerge := branchCfg.Merge.Short()
					b.Upstream = fmt.Sprintf("%s/%s", branchCfg.Remote, shortMerge)
				}
			}
		}

		branches = append(branches, b)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (s *GitService) CreateBranch(path, name, base string) error {
	return s.backend.CreateBranch(context.Background(), path, name, base)
}

func (s *GitService) DeleteBranch(path, name string, force bool) error {
	return s.backend.DeleteBranch(context.Background(), path, name)
}

func (s *GitService) RenameBranch(path, oldName, newName string) error {
	return s.backend.RenameBranch(context.Background(), path, oldName, newName)
}

func (s *GitService) GetBranchDescription(path, branch string) (string, error) {
	out, err := s.RunCommand(path, "config", fmt.Sprintf("branch.%s.description", branch))
	if err != nil {
		return "", nil
	}
	return out, nil
}

func (s *GitService) SetBranchDescription(path, branch, desc string) error {
	_, err := s.RunCommand(path, "config", fmt.Sprintf("branch.%s.description", branch), desc)
	return err
}

// GetBranchMetrics returns simple metrics: commit count
func (s *GitService) GetBranchMetrics(path, branch string) (map[string]int, error) {
	commits, err := s.backend.GetCommitsBetween(context.Background(), path, "", branch)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"commit_count": len(commits),
	}, nil
}
