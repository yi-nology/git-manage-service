package rag

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
)

type Service struct {
	store  *VectorStore
	client *EmbeddingClient
	gitSvc *git.GitService
}

var defaultService *Service

func InitService() {
	client := NewEmbeddingClientFromDB()
	gitSvc := git.NewGitService()
	defaultService = &Service{
		store:  DefaultStore(),
		client: client,
		gitSvc: gitSvc,
	}
	if client != nil {
		log.Printf("[RAG] Service initialized (provider_type=%s, model=%s)", client.ProviderType(), client.Model())
	} else {
		log.Printf("[RAG] Service initialized (no embedding client available)")
	}
}

func DefaultService() *Service {
	if defaultService == nil {
		InitService()
	}
	return defaultService
}

func (s *Service) IsAvailable() bool {
	return s.client != nil
}

func (s *Service) IndexRepo(ctx context.Context, repoKey string) (*IndexResult, error) {
	start := time.Now()

	if !s.IsAvailable() {
		return &IndexResult{RepoKey: repoKey, Error: "embedding client not configured"}, nil
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return &IndexResult{RepoKey: repoKey, Error: fmt.Sprintf("repo not found: %v", err)}, nil
	}

	tree, err := s.gitSvc.GetTree(repo.Path, "HEAD", "", true)
	if err != nil {
		return &IndexResult{RepoKey: repoKey, Error: fmt.Sprintf("get tree: %v", err)}, nil
	}

	sourceFiles := filterSourceFiles(tree)
	if len(sourceFiles) == 0 {
		return &IndexResult{RepoKey: repoKey, Error: "no source files found"}, nil
	}

	maxFiles := 200
	if len(sourceFiles) > maxFiles {
		sourceFiles = sourceFiles[:maxFiles]
	}

	var allChunks []*Chunk
	fileCount := 0
	for _, entry := range sourceFiles {
		blob, err := s.gitSvc.GetBlob(repo.Path, "HEAD", entry.Path)
		if err != nil || blob.IsBinary {
			continue
		}
		content := blob.Content
		if len(content) > 50000 {
			content = content[:50000]
		}
		chunks := ChunkFile(entry.Path, content, 800)
		allChunks = append(allChunks, chunks...)
		fileCount++
	}

	if len(allChunks) == 0 {
		return &IndexResult{RepoKey: repoKey, Error: "no chunks generated"}, nil
	}

	texts := make([]string, len(allChunks))
	for i, c := range allChunks {
		texts[i] = truncateForEmbedding(c.Content, 2000)
	}

	embeddings, err := s.client.Embed(ctx, texts)
	if err != nil {
		return &IndexResult{RepoKey: repoKey, Error: fmt.Sprintf("embedding failed: %v", err)}, nil
	}

	vectors := make([]*Vector, len(allChunks))
	for i, chunk := range allChunks {
		vec := &Vector{
			ID:        fmt.Sprintf("%s:%d", chunk.FilePath, chunk.StartLine),
			FilePath:  chunk.FilePath,
			Content:   chunk.Content,
			StartLine: chunk.StartLine,
			EndLine:   chunk.EndLine,
		}
		if i < len(embeddings) {
			vec.Values = embeddings[i]
		}
		vectors[i] = vec
	}

	s.store.Index(repoKey, vectors)

	return &IndexResult{
		RepoKey:    repoKey,
		ChunkCount: len(vectors),
		FileCount:  fileCount,
		Duration:   time.Since(start),
	}, nil
}

func (s *Service) IndexRemoteRepo(ctx context.Context, p provider.Provider, owner, repo, branch string) (*IndexResult, error) {
	start := time.Now()

	if !s.IsAvailable() {
		return &IndexResult{Error: "embedding client not configured"}, nil
	}

	repoKey := fmt.Sprintf("remote:%s/%s/%s", string(p.Platform()), owner, repo)

	files, err := p.GetCRFiles(ctx, owner, repo, 0)
	if err != nil || len(files) == 0 {
		return &IndexResult{RepoKey: repoKey, Error: fmt.Sprintf("list remote files: %v", err)}, nil
	}

	var allChunks []*Chunk
	fileCount := 0
	ref := branch
	if ref == "" {
		ref = "HEAD"
	}

	for _, f := range files {
		if f.IsBinary || f.Diff == "" {
			continue
		}
		path := f.NewPath
		if path == "" {
			path = f.OldPath
		}
		chunks := ChunkFile(path, f.Diff, 800)
		allChunks = append(allChunks, chunks...)
		fileCount++
	}

	if len(allChunks) == 0 {
		return &IndexResult{RepoKey: repoKey, Error: "no chunks from remote diff"}, nil
	}

	texts := make([]string, len(allChunks))
	for i, c := range allChunks {
		texts[i] = truncateForEmbedding(c.Content, 2000)
	}

	embeddings, err := s.client.Embed(ctx, texts)
	if err != nil {
		return &IndexResult{RepoKey: repoKey, Error: fmt.Sprintf("embedding failed: %v", err)}, nil
	}

	vectors := make([]*Vector, len(allChunks))
	for i, chunk := range allChunks {
		vec := &Vector{
			ID:        fmt.Sprintf("%s:%d", chunk.FilePath, chunk.StartLine),
			FilePath:  chunk.FilePath,
			Content:   chunk.Content,
			StartLine: chunk.StartLine,
			EndLine:   chunk.EndLine,
		}
		if i < len(embeddings) {
			vec.Values = embeddings[i]
		}
		vectors[i] = vec
	}

	s.store.Index(repoKey, vectors)

	return &IndexResult{
		RepoKey:    repoKey,
		ChunkCount: len(vectors),
		FileCount:  fileCount,
		Duration:   time.Since(start),
	}, nil
}

func (s *Service) Retrieve(ctx context.Context, repoKey string, changedFiles []string, topK int) ([]*SearchResult, error) {
	if !s.IsAvailable() {
		return nil, nil
	}

	if !s.store.HasIndex(repoKey) {
		return nil, nil
	}

	if topK <= 0 {
		topK = 5
	}

	var queries []string
	for _, f := range changedFiles {
		parts := strings.Split(f, "/")
		name := parts[len(parts)-1]
		queries = append(queries, "code context for file "+name+" in path "+f)
	}
	if len(queries) == 0 {
		return nil, nil
	}

	combined := strings.Join(queries, ". ")
	queryVec, err := s.client.EmbedQuery(ctx, combined)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	return s.store.Search(repoKey, queryVec, topK, 0.3), nil
}

func (s *Service) RetrieveForReview(ctx context.Context, repoID uint, files []string) (string, error) {
	if !s.IsAvailable() || len(files) == 0 {
		return "", nil
	}

	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return "", nil
	}

	repoKey := repo.Key
	_, idxErr := s.IndexRepo(ctx, repoKey)
	if idxErr != nil {
		log.Printf("[RAG] Index error (non-fatal): %v", idxErr)
	}

	results, err := s.Retrieve(ctx, repoKey, files, 5)
	if err != nil || len(results) == 0 {
		return "", nil
	}

	return FormatContextForPrompt(results, 6000), nil
}

func filterSourceFiles(tree []git.TreeEntry) []git.TreeEntry {
	var source []git.TreeEntry
	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".java": true, ".kt": true, ".rs": true, ".c": true, ".cpp": true, ".h": true,
		".cs": true, ".rb": true, ".php": true, ".swift": true, ".scala": true,
		".vue": true, ".svelte": true, ".sql": true, ".proto": true, ".graphql": true,
		".yaml": true, ".yml": true, ".toml": true, ".json": true, ".xml": true,
		".md": true, ".txt": true, ".sh": true, ".bash": true, ".zsh": true,
		".dockerfile": true, ".makefile": true,
	}

	skipDirs := map[string]bool{
		"vendor": true, "node_modules": true, "dist": true, "build": true,
		".git": true, "__pycache__": true, ".venv": true, "venv": true,
		"target": true, ".gradle": true, ".idea": true, ".vscode": true,
	}

	for _, entry := range tree {
		if entry.Type != "file" {
			continue
		}

		pathParts := strings.Split(entry.Path, "/")
		skip := false
		for _, p := range pathParts[:len(pathParts)-1] {
			if skipDirs[p] {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		for ext := range sourceExts {
			if strings.HasSuffix(strings.ToLower(entry.Path), ext) {
				source = append(source, entry)
				break
			}
		}

		name := entry.Path
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		switch strings.ToLower(name) {
		case "makefile", "dockerfile", "jenkinsfile", "vagrantfile", "readme", "license":
			source = append(source, entry)
		}
	}
	return source
}

func truncateForEmbedding(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen]
}
