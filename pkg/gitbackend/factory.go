package gitbackend

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/pkg/configs"
)

func NewGitBackendFromConfig(cfg configs.MirrorConfig) (GitBackend, error) {
	switch cfg.GitBackend {
	case "native":
		backend, err := NewNativeGitBackend()
		if err != nil {
			return nil, fmt.Errorf("native git backend unavailable, falling back: %w", err)
		}
		return backend, nil
	case "gogit":
		return NewGoGitBackend(), nil
	default:
		backend, err := NewNativeGitBackend()
		if err != nil {
			return NewGoGitBackend(), nil
		}
		return backend, nil
	}
}
