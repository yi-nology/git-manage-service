package rag

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type Vector struct {
	ID        string
	FilePath  string
	Content   string
	StartLine int
	EndLine   int
	Values    []float64
}

type VectorStore struct {
	mu         sync.RWMutex
	vectors    map[string][]*Vector
	maxPerRepo int
}

var defaultStore = NewVectorStore(5000)

func DefaultStore() *VectorStore {
	return defaultStore
}

func NewVectorStore(maxPerRepo int) *VectorStore {
	return &VectorStore{
		vectors:    make(map[string][]*Vector),
		maxPerRepo: maxPerRepo,
	}
}

func (s *VectorStore) Index(repoKey string, vectors []*Vector) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors[repoKey] = vectors

	if len(s.vectors[repoKey]) > s.maxPerRepo {
		s.vectors[repoKey] = s.vectors[repoKey][:s.maxPerRepo]
	}
}

func (s *VectorStore) Search(repoKey string, query []float64, topK int, minScore float64) []*SearchResult {
	s.mu.RLock()
	vectors, ok := s.vectors[repoKey]
	s.mu.RUnlock()

	if !ok || len(vectors) == 0 {
		return nil
	}

	results := make([]*SearchResult, 0, len(vectors))
	for _, v := range vectors {
		score := cosineSimilarity(query, v.Values)
		if score >= minScore {
			results = append(results, &SearchResult{
				Vector: v,
				Score:  score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

func (s *VectorStore) HasIndex(repoKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vectors, ok := s.vectors[repoKey]
	return ok && len(vectors) > 0
}

func (s *VectorStore) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := make(map[string]int, len(s.vectors))
	for k, v := range s.vectors {
		stats[k] = len(v)
	}
	return stats
}

type SearchResult struct {
	Vector *Vector
	Score  float64
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

type IndexResult struct {
	RepoKey    string
	ChunkCount int
	FileCount  int
	Duration   time.Duration
	Error      string
}

func FormatContextForPrompt(results []*SearchResult, maxChars int) string {
	var b strings.Builder
	remaining := maxChars
	for i, r := range results {
		if remaining <= 0 {
			break
		}
		if i > 0 {
			b.WriteString("\n---\n")
			remaining -= 4
		}
		header := fmt.Sprintf("File: %s (lines %d-%d, relevance: %.2f)\n", r.Vector.FilePath, r.Vector.StartLine, r.Vector.EndLine, r.Score)
		if len(header) > remaining {
			break
		}
		b.WriteString(header)
		remaining -= len(header)

		content := r.Vector.Content
		if len(content) > remaining {
			content = content[:remaining]
		}
		b.WriteString("```\n")
		b.WriteString(content)
		b.WriteString("\n```\n")
		remaining -= len(content) + 14
	}
	return b.String()
}
